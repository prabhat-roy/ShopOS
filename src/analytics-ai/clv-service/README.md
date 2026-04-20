# clv-service

> Customer Lifetime Value prediction and segmentation for the ShopOS analytics-ai domain.

## Overview

The clv-service computes Customer Lifetime Value (CLV) scores from RFM (Recency, Frequency, Monetary) features derived from order history and behavioural signals. Customers are segmented into Platinum, Gold, Silver, and Bronze tiers. Scores are persisted in PostgreSQL and served via HTTP and gRPC.

## Architecture

```mermaid
graph LR
    RS([recommendation-service]) -->|gRPC GetCustomerTier| CLV[clv-service]
    PS([personalization-service]) -->|gRPC GetCLVScore| CLV
    CLV -->|Read/Write scores| PG[(PostgreSQL)]
    CLV -->|Segment updates| Kafka[[Kafka]]
    OrdSvc([order-service]) -->|Order history feed| CLV
```

## Tech Stack

| Component | Technology |
|---|---|
| Language | Python 3.13 |
| Framework | FastAPI + uvicorn |
| Database | PostgreSQL (psycopg2) |
| ML | scikit-learn (RFM model) |
| Containerization | Docker (slim runtime) |

## CLV Tiers

| Tier | CLV Score | Description |
|---|---|---|
| Platinum | >= 80 | Top 5% — highest value customers |
| Gold | 60–79 | High value customers |
| Silver | 35–59 | Mid-tier customers |
| Bronze | < 35 | Entry-level customers |

## API Endpoints

| Endpoint | Method | Description |
|---|---|---|
| `/healthz` | GET | Liveness probe |
| `/clv/predict` | POST | Predict CLV score and tier for a customer |
| `/clv/tiers` | GET | Return tier thresholds |
| `/docs` | GET | Swagger UI |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `8195` | HTTP port |
| `GRPC_PORT` | `50190` | gRPC port |
| `DATABASE_URL` | `postgresql://clv_user:clv_pass@localhost:5432/clv_db` | PostgreSQL DSN |
| `CLV_PREDICTION_HORIZON_DAYS` | `365` | Prediction horizon for revenue forecast |
| `LOG_LEVEL` | `info` | Logging verbosity |

## Running Locally

```bash
docker-compose up clv-service
```

## Health Check

`GET /healthz` → `{"status":"ok"}`
