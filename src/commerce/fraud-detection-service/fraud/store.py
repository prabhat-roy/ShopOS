"""asyncpg-backed persistence for fraud check results."""

from __future__ import annotations

import json
from typing import Optional

import asyncpg

from fraud.models import FraudCheckResult

_MIGRATION_SQL = """
CREATE TABLE IF NOT EXISTS fraud_results (
    order_id        TEXT PRIMARY KEY,
    risk_score      INTEGER NOT NULL,
    risk_level      TEXT NOT NULL,
    decision        TEXT NOT NULL,
    signals         JSONB NOT NULL DEFAULT '[]',
    checked_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_fraud_results_decision ON fraud_results (decision);
CREATE INDEX IF NOT EXISTS idx_fraud_results_checked_at ON fraud_results (checked_at DESC);
"""


class FraudStore:
    def __init__(self, pool: asyncpg.Pool) -> None:
        self._pool = pool

    @classmethod
    async def create(cls, database_url: str) -> "FraudStore":
        pool = await asyncpg.create_pool(database_url, min_size=2, max_size=10)
        store = cls(pool)
        await store._migrate()
        return store

    async def _migrate(self) -> None:
        async with self._pool.acquire() as conn:
            await conn.execute(_MIGRATION_SQL)

    async def close(self) -> None:
        await self._pool.close()

    async def save_result(self, result: FraudCheckResult) -> None:
        sql = """
            INSERT INTO fraud_results
                (order_id, risk_score, risk_level, decision, signals, checked_at)
            VALUES ($1, $2, $3, $4, $5::jsonb, $6)
            ON CONFLICT (order_id) DO UPDATE SET
                risk_score  = EXCLUDED.risk_score,
                risk_level  = EXCLUDED.risk_level,
                decision    = EXCLUDED.decision,
                signals     = EXCLUDED.signals,
                checked_at  = EXCLUDED.checked_at
        """
        async with self._pool.acquire() as conn:
            await conn.execute(
                sql,
                result.order_id,
                result.risk_score,
                result.risk_level,
                result.decision,
                json.dumps(result.signals),
                result.checked_at,
            )

    async def get_result(self, order_id: str) -> Optional[FraudCheckResult]:
        sql = """
            SELECT order_id, risk_score, risk_level, decision, signals, checked_at
            FROM fraud_results
            WHERE order_id = $1
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(sql, order_id)
        if row is None:
            return None
        return FraudCheckResult(
            order_id=row["order_id"],
            risk_score=row["risk_score"],
            risk_level=row["risk_level"],
            decision=row["decision"],
            signals=json.loads(row["signals"]),
            checked_at=row["checked_at"],
        )

    async def list_flagged(self, limit: int = 50) -> list[FraudCheckResult]:
        sql = """
            SELECT order_id, risk_score, risk_level, decision, signals, checked_at
            FROM fraud_results
            WHERE decision IN ('review', 'decline')
            ORDER BY checked_at DESC
            LIMIT $1
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql, limit)
        return [
            FraudCheckResult(
                order_id=r["order_id"],
                risk_score=r["risk_score"],
                risk_level=r["risk_level"],
                decision=r["decision"],
                signals=json.loads(r["signals"]),
                checked_at=r["checked_at"],
            )
            for r in rows
        ]
