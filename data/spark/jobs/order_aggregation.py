"""Spark job: real-time order aggregation from Kafka → ClickHouse.

Consumes commerce.order.placed events from Kafka, computes micro-batch
revenue aggregates, writes to ClickHouse for Grafana dashboards.
"""
from pyspark.sql import SparkSession
from pyspark.sql.functions import (
    col, from_json, sum as spark_sum, count, avg,
    window, to_timestamp, lit,
)
from pyspark.sql.types import (
    StructType, StructField, StringType, DoubleType,
    IntegerType, TimestampType,
)

ORDER_SCHEMA = StructType([
    StructField("order_id",     StringType(),  False),
    StructField("user_id",      StringType(),  False),
    StructField("total_amount", DoubleType(),  False),
    StructField("currency",     StringType(),  True),
    StructField("status",       StringType(),  False),
    StructField("item_count",   IntegerType(), True),
    StructField("created_at",   StringType(),  False),
])


def main() -> None:
    spark = (
        SparkSession.builder
        .appName("ShopOS-OrderAggregation")
        .config("spark.streaming.stopGracefullyOnShutdown", "true")
        .config("spark.sql.streaming.schemaInference", "true")
        .getOrCreate()
    )
    spark.sparkContext.setLogLevel("WARN")

    kafka_df = (
        spark.readStream
        .format("kafka")
        .option("kafka.bootstrap.servers", "kafka.streaming.svc.cluster.local:9092")
        .option("subscribe", "commerce.order.placed")
        .option("startingOffsets", "latest")
        .option("kafka.security.protocol", "PLAINTEXT")
        .load()
    )

    orders = (
        kafka_df
        .selectExpr("CAST(value AS STRING) AS json_str", "timestamp")
        .select(
            from_json(col("json_str"), ORDER_SCHEMA).alias("data"),
            col("timestamp"),
        )
        .select("data.*", col("timestamp").alias("kafka_ts"))
        .withColumn("created_at", to_timestamp(col("created_at")))
    )

    agg = (
        orders
        .withWatermark("created_at", "5 minutes")
        .groupBy(window(col("created_at"), "1 minute").alias("window"))
        .agg(
            count("order_id").alias("order_count"),
            spark_sum("total_amount").alias("revenue"),
            avg("total_amount").alias("avg_order_value"),
            spark_sum("item_count").alias("total_items"),
        )
        .select(
            col("window.start").alias("window_start"),
            col("window.end").alias("window_end"),
            col("order_count"),
            col("revenue"),
            col("avg_order_value"),
            col("total_items"),
            lit("1m").alias("granularity"),
        )
    )

    query = (
        agg.writeStream
        .outputMode("update")
        .format("console")
        .option("truncate", "false")
        .option("checkpointLocation", "/tmp/checkpoints/order-aggregation")
        .trigger(processingTime="30 seconds")
        .start()
    )

    query.awaitTermination()


if __name__ == "__main__":
    main()
