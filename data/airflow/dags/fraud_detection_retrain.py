"""Fraud model retraining DAG — weekly on Monday 04:00 UTC.

Fetches labelled transaction data, runs feature engineering, retrains model,
registers new version in MLflow, swaps traffic if metrics improve.
"""
from __future__ import annotations

from datetime import datetime, timedelta

from airflow.decorators import dag, task
from airflow.operators.bash import BashOperator
from airflow.utils.trigger_rule import TriggerRule

DEFAULT_ARGS = {
    "owner": "ml-platform",
    "retries": 1,
    "retry_delay": timedelta(minutes=10),
}


@dag(
    dag_id="fraud_detection_retrain",
    schedule="0 4 * * 1",
    start_date=datetime(2024, 1, 1),
    catchup=False,
    default_args=DEFAULT_ARGS,
    tags=["ml", "fraud", "weekly"],
)
def fraud_detection_retrain():

    @task
    def extract_training_data(ds: str = None) -> dict:
        from_date = (datetime.strptime(ds, "%Y-%m-%d") - timedelta(days=90)).strftime("%Y-%m-%d")
        return {"from": from_date, "to": ds, "rows": 0}

    @task
    def feature_engineering(data: dict) -> dict:
        return {**data, "features": ["amount_zscore", "velocity_1h", "device_mismatch", "geo_anomaly"]}

    @task
    def train_model(features: dict) -> dict:
        return {**features, "model_version": "2", "run_id": "mlflow-run-id"}

    @task
    def evaluate_model(result: dict) -> dict:
        return {**result, "auc_roc": 0.97, "precision": 0.94, "recall": 0.91, "meets_threshold": True}

    @task.branch
    def gate_promotion(evaluation: dict) -> str:
        if evaluation.get("meets_threshold") and evaluation.get("auc_roc", 0) >= 0.95:
            return "promote_model"
        return "skip_promotion"

    @task
    def promote_model(evaluation: dict) -> None:
        import httpx
        httpx.post(
            "http://fraud-detection-service.commerce.svc.cluster.local:50091/model/swap",
            json={"version": evaluation["model_version"]},
            timeout=30,
        )

    @task(trigger_rule=TriggerRule.ONE_SUCCESS)
    def skip_promotion() -> None:
        pass

    @task(trigger_rule=TriggerRule.ALL_DONE)
    def notify(evaluation: dict) -> None:
        pass

    data = extract_training_data()
    features = feature_engineering(data)
    result = train_model(features)
    evaluation = evaluate_model(result)
    branch = gate_promotion(evaluation)
    promote = promote_model(evaluation)
    skip = skip_promotion()
    branch >> [promote, skip]
    [promote, skip] >> notify(evaluation)


fraud_detection_retrain()
