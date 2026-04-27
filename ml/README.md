# ML â€” ShopOS

Machine learning platform tooling and configurations.

## Directory Structure

```
ml/
â”œâ”€â”€ mlflow/             â† MLflow experiment tracking, model registry, artifact store
â””â”€â”€ charts/             â† Helm charts for ML platform components
```

## Deployed Stack

| Component | Version | Role |
|---|---|---|
| MLflow | 2.16 | ML experiment tracking, model versioning, artifact registry |
| Weaviate | 1.26 | Vector database â€” semantic search, recommendation embeddings |
| Neo4j | 5.23 | Graph database â€” product recommendation graph traversal |

> Weaviate and Neo4j configs are in `databases/weaviate/` and `databases/neo4j/` respectively.

## ML Services (Analytics/AI Domain)

| Service | Role |
|---|---|
| `recommendation-service` | Collaborative filtering + Neo4j graph-based product recommendations |
| `sentiment-analysis-service` | NLP on product reviews; outputs positive/negative/neutral scores |
| `price-optimization-service` | ML-driven dynamic pricing suggestions |
| `ml-feature-store` | Centralised feature store; features shared across models |
| `personalization-service` | User-specific product ranking and homepage personalisation |
| `data-pipeline-service` | ETL from Cassandra â†’ feature store â†’ model training data |
| `clv-service` | Customer Lifetime Value prediction |
| `attribution-service` | Multi-touch marketing attribution modelling |
| `search-analytics-service` | Search relevance tuning and query analytics |

## MLflow Setup

MLflow tracks all model training runs across services. Configuration:

- Tracking server: `http://mlflow:5000`
- Artifact store: MinIO (`s3://mlflow-artifacts`)
- Backend store: PostgreSQL (`mlflow` database)

Services that train models log experiments via the MLflow Python SDK:

```python
import mlflow

mlflow.set_tracking_uri("http://mlflow:5000")
with mlflow.start_run():
    mlflow.log_params({"learning_rate": 0.01, "epochs": 100})
    mlflow.log_metric("rmse", 0.42)
    mlflow.sklearn.log_model(model, "model")
```

## Model Serving

Models are served via the `recommendation-service` and `personalization-service` gRPC APIs. The services load the latest production model from the MLflow Model Registry at startup.

## Future (Phase 5)

- Feast / Hopsworks: Managed feature store with real-time feature serving
- Argo Workflows: Automated model training pipelines triggered by data drift
- LLM Service: FastAPI wrapper around an OpenAI-compatible LLM API
- RAG Pipeline: Weaviate vector search + LLM for semantic product Q&A

## References

- [Analytics/AI Domain Services](../docs/architecture/domain-map.md#10-analytics--ai-domain)
- [Database Strategy â€” Weaviate / Neo4j](../docs/architecture/database-strategy.md)
