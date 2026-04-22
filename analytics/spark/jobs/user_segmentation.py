"""Spark batch job: user segmentation for personalization and marketing.

Reads orders + page views from Postgres/ClickHouse, computes RFM scores,
segments users into cohorts, writes segments to Postgres for personalization-service.
"""
from pyspark.sql import SparkSession
from pyspark.sql.functions import (
    col, count, sum as spark_sum, max as spark_max,
    datediff, current_date, ntile, when, concat_ws,
)
from pyspark.sql.window import Window


def main() -> None:
    spark = (
        SparkSession.builder
        .appName("ShopOS-UserSegmentation")
        .getOrCreate()
    )

    pg_opts = {
        "url": "jdbc:postgresql://pgbouncer.databases.svc.cluster.local:5432/shopos",
        "user": "shopos_spark",
        "password": "INJECTED_BY_ENV",
        "driver": "org.postgresql.Driver",
    }

    orders = (
        spark.read.format("jdbc")
        .options(**pg_opts, dbtable="analytics.fct_orders")
        .load()
        .filter(col("is_completed") == True)
        .filter(col("ordered_at") >= "2024-01-01")
    )

    # RFM: Recency, Frequency, Monetary
    rfm = orders.groupBy("user_id").agg(
        datediff(current_date(), spark_max("ordered_at")).alias("recency_days"),
        count("order_id").alias("frequency"),
        spark_sum("total_usd").alias("monetary"),
    )

    w = Window.orderBy(col("recency_days").asc())
    rfm_scored = rfm.withColumn("r_score", ntile(5).over(Window.orderBy(col("recency_days").asc()))) \
                    .withColumn("f_score", ntile(5).over(Window.orderBy(col("frequency").desc()))) \
                    .withColumn("m_score", ntile(5).over(Window.orderBy(col("monetary").desc())))

    segmented = rfm_scored.withColumn(
        "segment",
        when((col("r_score") >= 4) & (col("f_score") >= 4) & (col("m_score") >= 4), "champions")
        .when((col("r_score") >= 3) & (col("f_score") >= 3), "loyal")
        .when((col("r_score") >= 4) & (col("f_score") <= 2), "new_customers")
        .when((col("r_score") <= 2) & (col("f_score") >= 3), "at_risk")
        .when((col("r_score") <= 2) & (col("f_score") <= 2), "lost")
        .otherwise("potential")
    ).withColumn(
        "rfm_label",
        concat_ws("-",
            col("r_score").cast("string"),
            col("f_score").cast("string"),
            col("m_score").cast("string"),
        )
    )

    # Write back to Postgres for personalization-service to consume
    (
        segmented
        .select("user_id", "segment", "rfm_label", "recency_days", "frequency", "monetary",
                "r_score", "f_score", "m_score")
        .write
        .format("jdbc")
        .options(**pg_opts, dbtable="analytics.user_segments")
        .mode("overwrite")
        .save()
    )

    print(f"Segmented {segmented.count()} users")
    spark.stop()


if __name__ == "__main__":
    main()
