from __future__ import annotations

from datetime import datetime, timezone
from typing import Optional

import asyncpg

from .models import PriceOptimizationResult, PricingRule


CREATE_RESULTS_TABLE = """
CREATE TABLE IF NOT EXISTS price_optimization_results (
    id              BIGSERIAL PRIMARY KEY,
    product_id      TEXT        NOT NULL,
    current_price   NUMERIC     NOT NULL,
    suggested_price NUMERIC     NOT NULL,
    demand_change   NUMERIC     NOT NULL,
    revenue_change  NUMERIC     NOT NULL,
    margin          NUMERIC     NOT NULL,
    confidence      NUMERIC     NOT NULL,
    reasoning       TEXT        NOT NULL,
    generated_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_price_results_product_id
    ON price_optimization_results (product_id, generated_at DESC);
"""

CREATE_RULES_TABLE = """
CREATE TABLE IF NOT EXISTS pricing_rules (
    product_id    TEXT    PRIMARY KEY,
    min_price     NUMERIC NOT NULL,
    max_price     NUMERIC NOT NULL,
    target_margin NUMERIC NOT NULL,
    active        BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at    TIMESTAMPTZ NOT NULL
);
"""


class AsyncPgStore:
    def __init__(self, database_url: str) -> None:
        self._database_url = database_url
        self._pool: Optional[asyncpg.Pool] = None

    async def connect(self) -> None:
        self._pool = await asyncpg.create_pool(self._database_url, min_size=2, max_size=10)
        async with self._pool.acquire() as conn:
            await conn.execute(CREATE_RESULTS_TABLE)
            await conn.execute(CREATE_RULES_TABLE)

    async def disconnect(self) -> None:
        if self._pool:
            await self._pool.close()

    # ------------------------------------------------------------------
    # Results
    # ------------------------------------------------------------------

    async def save_result(self, result: PriceOptimizationResult) -> None:
        sql = """
            INSERT INTO price_optimization_results
                (product_id, current_price, suggested_price,
                 demand_change, revenue_change, margin, confidence,
                 reasoning, generated_at)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
        """
        async with self._pool.acquire() as conn:
            await conn.execute(
                sql,
                result.productId,
                result.currentPrice,
                result.suggestedPrice,
                result.expectedDemandChange,
                result.expectedRevenueChange,
                result.marginAtSuggestedPrice,
                result.confidence,
                result.reasoning,
                result.generatedAt,
            )

    async def get_result(self, product_id: str) -> Optional[PriceOptimizationResult]:
        sql = """
            SELECT product_id, current_price, suggested_price,
                   demand_change, revenue_change, margin, confidence,
                   reasoning, generated_at
            FROM   price_optimization_results
            WHERE  product_id = $1
            ORDER  BY generated_at DESC
            LIMIT  1
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(sql, product_id)
        if row is None:
            return None
        return _row_to_result(row)

    async def list_results(self, limit: int = 100) -> list[PriceOptimizationResult]:
        sql = """
            SELECT DISTINCT ON (product_id)
                   product_id, current_price, suggested_price,
                   demand_change, revenue_change, margin, confidence,
                   reasoning, generated_at
            FROM   price_optimization_results
            ORDER  BY product_id, generated_at DESC
            LIMIT  $1
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql, limit)
        return [_row_to_result(r) for r in rows]

    # ------------------------------------------------------------------
    # Rules
    # ------------------------------------------------------------------

    async def save_rule(self, rule: PricingRule) -> None:
        sql = """
            INSERT INTO pricing_rules (product_id, min_price, max_price,
                                       target_margin, active, updated_at)
            VALUES ($1,$2,$3,$4,$5,$6)
            ON CONFLICT (product_id)
            DO UPDATE SET
                min_price     = EXCLUDED.min_price,
                max_price     = EXCLUDED.max_price,
                target_margin = EXCLUDED.target_margin,
                active        = EXCLUDED.active,
                updated_at    = EXCLUDED.updated_at
        """
        async with self._pool.acquire() as conn:
            await conn.execute(
                sql,
                rule.productId,
                rule.minPrice,
                rule.maxPrice,
                rule.targetMargin,
                rule.active,
                datetime.now(timezone.utc),
            )

    async def get_rule(self, product_id: str) -> Optional[PricingRule]:
        sql = """
            SELECT product_id, min_price, max_price, target_margin, active
            FROM   pricing_rules
            WHERE  product_id = $1
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(sql, product_id)
        if row is None:
            return None
        return _row_to_rule(row)

    async def list_rules(self) -> list[PricingRule]:
        sql = """
            SELECT product_id, min_price, max_price, target_margin, active
            FROM   pricing_rules
            ORDER  BY product_id
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql)
        return [_row_to_rule(r) for r in rows]

    async def delete_rule(self, product_id: str) -> bool:
        sql = "DELETE FROM pricing_rules WHERE product_id = $1"
        async with self._pool.acquire() as conn:
            result = await conn.execute(sql, product_id)
        return result.endswith("1")


def _row_to_result(row: asyncpg.Record) -> PriceOptimizationResult:
    return PriceOptimizationResult(
        productId=row["product_id"],
        currentPrice=float(row["current_price"]),
        suggestedPrice=float(row["suggested_price"]),
        expectedDemandChange=float(row["demand_change"]),
        expectedRevenueChange=float(row["revenue_change"]),
        marginAtSuggestedPrice=float(row["margin"]),
        confidence=float(row["confidence"]),
        reasoning=row["reasoning"],
        generatedAt=row["generated_at"],
    )


def _row_to_rule(row: asyncpg.Record) -> PricingRule:
    return PricingRule(
        productId=row["product_id"],
        minPrice=float(row["min_price"]),
        maxPrice=float(row["max_price"]),
        targetMargin=float(row["target_margin"]),
        active=row["active"],
    )
