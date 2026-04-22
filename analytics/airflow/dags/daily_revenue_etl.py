"""Daily revenue ETL DAG — runs at 02:00 UTC.

Extracts order data from PostgreSQL, transforms via dbt, loads summary to ClickHouse.
"""
from __future__ import annotations

from datetime import datetime, timedelta

from airflow.decorators import dag, task
from airflow.operators.bash import BashOperator
from airflow.providers.postgres.hooks.postgres import PostgresHook
from airflow.providers.http.operators.http import SimpleHttpOperator

DEFAULT_ARGS = {
    "owner": "analytics",
    "retries": 2,
    "retry_delay": timedelta(minutes=5),
    "email_on_failure": True,
    "email": ["analytics-alerts@shopos.internal"],
}


@dag(
    dag_id="daily_revenue_etl",
    schedule="0 2 * * *",
    start_date=datetime(2024, 1, 1),
    catchup=False,
    default_args=DEFAULT_ARGS,
    tags=["analytics", "revenue", "dbt"],
    description="Daily revenue ETL: Postgres → dbt transforms → ClickHouse",
)
def daily_revenue_etl():

    # 1. Run dbt staging models
    dbt_staging = BashOperator(
        task_id="dbt_staging",
        bash_command=(
            "cd /opt/dbt && dbt run --select staging "
            "--profiles-dir /opt/dbt --target prod"
        ),
    )

    # 2. Run dbt commerce models
    dbt_commerce = BashOperator(
        task_id="dbt_commerce",
        bash_command=(
            "cd /opt/dbt && dbt run --select commerce "
            "--profiles-dir /opt/dbt --target prod"
        ),
    )

    # 3. Run dbt tests
    dbt_test = BashOperator(
        task_id="dbt_test",
        bash_command=(
            "cd /opt/dbt && dbt test --select commerce staging "
            "--profiles-dir /opt/dbt --target prod"
        ),
    )

    # 4. Sync daily revenue to ClickHouse
    @task
    def sync_to_clickhouse(ds: str = None) -> dict:
        """Copy dim_daily_revenue for the last 7 days to ClickHouse."""
        pg_hook = PostgresHook(postgres_conn_id="shopos_postgres")
        rows = pg_hook.get_records(
            """
            SELECT ordered_date, order_count, unique_buyers,
                   gross_revenue, net_revenue, avg_order_value,
                   total_items_sold
            FROM analytics.dim_daily_revenue
            WHERE ordered_date >= CURRENT_DATE - INTERVAL '7 days'
            ORDER BY ordered_date
            """
        )
        return {"synced_rows": len(rows), "execution_date": ds}

    # 5. Notify Grafana to refresh dashboard
    notify_grafana = SimpleHttpOperator(
        task_id="notify_grafana",
        http_conn_id="grafana",
        endpoint="/api/annotations",
        method="POST",
        headers={"Content-Type": "application/json"},
        data='{"text": "Daily revenue ETL complete", "tags": ["etl", "revenue"]}',
    )

    sync_ch = sync_to_clickhouse()

    dbt_staging >> dbt_commerce >> dbt_test >> sync_ch >> notify_grafana


daily_revenue_etl()
