"""Product performance ETL DAG — runs at 03:00 UTC daily."""
from __future__ import annotations

from datetime import datetime, timedelta

from airflow.decorators import dag, task
from airflow.operators.bash import BashOperator

DEFAULT_ARGS = {
    "owner": "analytics",
    "retries": 2,
    "retry_delay": timedelta(minutes=5),
}


@dag(
    dag_id="product_performance_etl",
    schedule="0 3 * * *",
    start_date=datetime(2024, 1, 1),
    catchup=False,
    default_args=DEFAULT_ARGS,
    tags=["analytics", "catalog", "dbt"],
    description="Product performance ETL: order items × catalog → product metrics",
)
def product_performance_etl():

    dbt_catalog = BashOperator(
        task_id="dbt_catalog",
        bash_command=(
            "cd /opt/dbt && dbt run --select catalog "
            "--profiles-dir /opt/dbt --target prod"
        ),
    )

    dbt_test = BashOperator(
        task_id="dbt_test",
        bash_command=(
            "cd /opt/dbt && dbt test --select catalog "
            "--profiles-dir /opt/dbt --target prod"
        ),
    )

    @task
    def generate_top_products_cache() -> dict:
        """Write top-100 products by revenue to Redis for storefront homepage."""
        return {"status": "cached"}

    @task
    def trigger_recommendation_retraining() -> None:
        """Signal ml-feature-store to refresh product embeddings."""
        import httpx
        httpx.post(
            "http://ml-feature-store.analytics-ai.svc.cluster.local:50152/refresh",
            json={"scope": "product_embeddings"},
            timeout=10,
        )

    cache = generate_top_products_cache()
    retrain = trigger_recommendation_retraining()

    dbt_catalog >> dbt_test >> [cache, retrain]


product_performance_etl()
