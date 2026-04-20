from __future__ import annotations

import json
import logging
from datetime import datetime, timezone
from typing import Optional

import asyncpg

from forecast.models import ForecastResponse, SalesRecord

logger = logging.getLogger(__name__)


CREATE_SALES_RECORDS_TABLE = """
CREATE TABLE IF NOT EXISTS sales_records (
    id          BIGSERIAL PRIMARY KEY,
    product_id  TEXT        NOT NULL,
    sku         TEXT        NOT NULL,
    quantity    INTEGER     NOT NULL CHECK (quantity > 0),
    sale_date   DATE        NOT NULL,
    order_id    TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (order_id, product_id)
);
CREATE INDEX IF NOT EXISTS idx_sales_records_product_date
    ON sales_records (product_id, sale_date DESC);
"""

CREATE_FORECASTS_TABLE = """
CREATE TABLE IF NOT EXISTS forecasts (
    id                  BIGSERIAL PRIMARY KEY,
    product_id          TEXT           NOT NULL UNIQUE,
    forecasted_demand   DOUBLE PRECISION NOT NULL,
    average_daily_sales DOUBLE PRECISION NOT NULL,
    historical_days     INTEGER        NOT NULL,
    confidence          DOUBLE PRECISION NOT NULL,
    generated_at        TIMESTAMPTZ    NOT NULL,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
"""


class AsyncPgStore:
    def __init__(self, database_url: str) -> None:
        self._database_url = database_url
        self._pool: Optional[asyncpg.Pool] = None

    async def init(self) -> None:
        self._pool = await asyncpg.create_pool(
            dsn=self._database_url,
            min_size=2,
            max_size=10,
            command_timeout=30,
        )
        async with self._pool.acquire() as conn:
            await conn.execute(CREATE_SALES_RECORDS_TABLE)
            await conn.execute(CREATE_FORECASTS_TABLE)
        logger.info("Database pool initialised and tables verified.")

    async def close(self) -> None:
        if self._pool:
            await self._pool.close()
            logger.info("Database pool closed.")

    # ------------------------------------------------------------------
    # Sales records
    # ------------------------------------------------------------------

    async def save_sale(self, record: SalesRecord) -> None:
        assert self._pool is not None, "Pool not initialised"
        await self._pool.execute(
            """
            INSERT INTO sales_records (product_id, sku, quantity, sale_date, order_id)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (order_id, product_id) DO NOTHING
            """,
            record.productId,
            record.sku,
            record.quantity,
            record.saleDate,
            record.orderId,
        )

    async def get_sales_history(self, product_id: str, days: int) -> list[SalesRecord]:
        assert self._pool is not None, "Pool not initialised"
        rows = await self._pool.fetch(
            """
            SELECT product_id, sku, quantity, sale_date, order_id
            FROM   sales_records
            WHERE  product_id = $1
              AND  sale_date  >= CURRENT_DATE - ($2 || ' days')::INTERVAL
            ORDER  BY sale_date DESC
            """,
            product_id,
            str(days),
        )
        return [
            SalesRecord(
                productId=row["product_id"],
                sku=row["sku"],
                quantity=row["quantity"],
                saleDate=row["sale_date"],
                orderId=row["order_id"],
            )
            for row in rows
        ]

    # ------------------------------------------------------------------
    # Forecasts
    # ------------------------------------------------------------------

    async def save_forecast(self, forecast: ForecastResponse) -> None:
        assert self._pool is not None, "Pool not initialised"
        await self._pool.execute(
            """
            INSERT INTO forecasts
                (product_id, forecasted_demand, average_daily_sales, historical_days,
                 confidence, generated_at)
            VALUES ($1, $2, $3, $4, $5, $6)
            ON CONFLICT (product_id) DO UPDATE
                SET forecasted_demand   = EXCLUDED.forecasted_demand,
                    average_daily_sales = EXCLUDED.average_daily_sales,
                    historical_days     = EXCLUDED.historical_days,
                    confidence          = EXCLUDED.confidence,
                    generated_at        = EXCLUDED.generated_at
            """,
            forecast.productId,
            forecast.forecastedDemand,
            forecast.averageDailySales,
            forecast.historicalDays,
            forecast.confidence,
            forecast.generatedAt,
        )

    async def get_forecast(self, product_id: str) -> Optional[ForecastResponse]:
        assert self._pool is not None, "Pool not initialised"
        row = await self._pool.fetchrow(
            """
            SELECT product_id, forecasted_demand, average_daily_sales,
                   historical_days, confidence, generated_at
            FROM   forecasts
            WHERE  product_id = $1
            """,
            product_id,
        )
        if row is None:
            return None
        return ForecastResponse(
            productId=row["product_id"],
            forecastedDemand=row["forecasted_demand"],
            averageDailySales=row["average_daily_sales"],
            historicalDays=row["historical_days"],
            confidence=row["confidence"],
            generatedAt=row["generated_at"],
        )

    async def list_forecasts(self, limit: int = 50) -> list[ForecastResponse]:
        assert self._pool is not None, "Pool not initialised"
        rows = await self._pool.fetch(
            """
            SELECT product_id, forecasted_demand, average_daily_sales,
                   historical_days, confidence, generated_at
            FROM   forecasts
            ORDER  BY generated_at DESC
            LIMIT  $1
            """,
            limit,
        )
        return [
            ForecastResponse(
                productId=row["product_id"],
                forecastedDemand=row["forecasted_demand"],
                averageDailySales=row["average_daily_sales"],
                historicalDays=row["historical_days"],
                confidence=row["confidence"],
                generatedAt=row["generated_at"],
            )
            for row in rows
        ]
