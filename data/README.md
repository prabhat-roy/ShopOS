# Data Platform — ShopOS

The unified data platform — workflow orchestration, ELT, transformation, BI, semantic
layer, lineage, and quality. Sits between the OLTP services (Postgres, MongoDB, Redis)
and the analytical stores (ClickHouse, Weaviate, Neo4j) used by reporting + analytics-ai.

## Layout

| Subdir | Tool | Role |
|---|---|---|
| [airflow/](airflow/) | Apache Airflow | DAG-based workflow orchestration — daily ETL, fraud retrain, product performance |
| [dbt/](dbt/) | dbt | SQL transformations — staging, commerce, catalog models with incremental materializations |
| [spark/](spark/) | Apache Spark | Batch processing — order aggregation streaming, user RFM segmentation |
| [airbyte/](airbyte/) | Airbyte | ELT from 300+ sources (Stripe, Salesforce, Postgres replicas) into ClickHouse |
| [cube/](cube/) | Cube | Semantic layer + BI API on top of ClickHouse / Postgres |
| [metabase/](metabase/) | Metabase | Self-serve BI dashboards (alternative to Superset) |
| [lineage/openlineage/](lineage/openlineage/) | OpenLineage + Marquez | Data lineage tracking — Airflow → dbt → ClickHouse pipeline DAG |
| [quality/great-expectations/](quality/great-expectations/) | Great Expectations | Data quality assertion suites on orders, products, events |

## Pipeline

```
OLTP services
   │ (Debezium CDC)
   ▼
Kafka topics ──► Apache Flink (real-time aggregations)
   │                  │
   ▼                  ▼
Apache Airflow ──► dbt ──► ClickHouse / TimescaleDB
   │                          │
   │                          ▼
   ▼                       Cube + Metabase + Superset
LakeFS-versioned MinIO     │
   │                       ▼
   ▼                    Storefront facets / admin reports
analytics-ai (recommendation, fraud, sentiment, CLV)
```

## Related

- ML platform: [ml/](../ml/)
- Stream processing: [streaming/](../streaming/)
- Analytical databases: [databases/](../databases/)
