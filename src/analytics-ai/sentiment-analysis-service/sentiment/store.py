from __future__ import annotations

import logging
from typing import Optional

import asyncpg

from .models import AggregateStats, SentimentLabel, SentimentResult

logger = logging.getLogger(__name__)

_CREATE_TABLE_SQL = """
CREATE TABLE IF NOT EXISTS sentiment_results (
    id              BIGSERIAL PRIMARY KEY,
    entity_id       TEXT,
    entity_type     TEXT,
    text            TEXT         NOT NULL,
    label           TEXT         NOT NULL,
    score           DOUBLE PRECISION NOT NULL,
    positive_words  TEXT[]       NOT NULL DEFAULT '{}',
    negative_words  TEXT[]       NOT NULL DEFAULT '{}',
    analyzed_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sentiment_entity_id   ON sentiment_results (entity_id);
CREATE INDEX IF NOT EXISTS idx_sentiment_entity_type ON sentiment_results (entity_type);
CREATE INDEX IF NOT EXISTS idx_sentiment_label       ON sentiment_results (label);
"""


class AsyncPgStore:
    """PostgreSQL-backed persistence for sentiment analysis results."""

    def __init__(self, pool: asyncpg.Pool) -> None:
        self._pool = pool

    # ------------------------------------------------------------------
    # Initialisation
    # ------------------------------------------------------------------

    @classmethod
    async def create(cls, database_url: str) -> "AsyncPgStore":
        pool = await asyncpg.create_pool(database_url, min_size=2, max_size=10)
        store = cls(pool)
        await store._init_schema()
        return store

    async def _init_schema(self) -> None:
        async with self._pool.acquire() as conn:
            await conn.execute(_CREATE_TABLE_SQL)
        logger.info("sentiment_results table ready")

    async def close(self) -> None:
        await self._pool.close()

    # ------------------------------------------------------------------
    # Write
    # ------------------------------------------------------------------

    async def save_result(self, result: SentimentResult) -> int:
        sql = """
            INSERT INTO sentiment_results
                (entity_id, entity_type, text, label, score,
                 positive_words, negative_words, analyzed_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                sql,
                result.entityId,
                result.entityType,
                result.text,
                result.label.value,
                result.score,
                result.positiveWords,
                result.negativeWords,
                result.analyzedAt,
            )
        return row["id"]

    # ------------------------------------------------------------------
    # Read
    # ------------------------------------------------------------------

    async def get_result(self, entity_id: str) -> Optional[SentimentResult]:
        sql = """
            SELECT * FROM sentiment_results
            WHERE entity_id = $1
            ORDER BY analyzed_at DESC
            LIMIT 1
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(sql, entity_id)

        if row is None:
            return None
        return self._row_to_model(row)

    async def list_results(
        self,
        entity_type: Optional[str] = None,
        limit: int = 50,
    ) -> list[SentimentResult]:
        if entity_type:
            sql = """
                SELECT * FROM sentiment_results
                WHERE entity_type = $1
                ORDER BY analyzed_at DESC
                LIMIT $2
            """
            params = (entity_type, limit)
        else:
            sql = """
                SELECT * FROM sentiment_results
                ORDER BY analyzed_at DESC
                LIMIT $1
            """
            params = (limit,)

        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql, *params)

        return [self._row_to_model(r) for r in rows]

    async def get_aggregate_stats(
        self, entity_type: Optional[str] = None
    ) -> AggregateStats:
        if entity_type:
            sql = """
                SELECT
                    COUNT(*) FILTER (WHERE label = 'POSITIVE') AS positive,
                    COUNT(*) FILTER (WHERE label = 'NEGATIVE') AS negative,
                    COUNT(*) FILTER (WHERE label = 'NEUTRAL')  AS neutral,
                    COUNT(*)                                    AS total,
                    COALESCE(AVG(score), 0)                    AS avg_score
                FROM sentiment_results
                WHERE entity_type = $1
            """
            params = (entity_type,)
        else:
            sql = """
                SELECT
                    COUNT(*) FILTER (WHERE label = 'POSITIVE') AS positive,
                    COUNT(*) FILTER (WHERE label = 'NEGATIVE') AS negative,
                    COUNT(*) FILTER (WHERE label = 'NEUTRAL')  AS neutral,
                    COUNT(*)                                    AS total,
                    COALESCE(AVG(score), 0)                    AS avg_score
                FROM sentiment_results
            """
            params = ()

        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(sql, *params)

        return AggregateStats(
            entityType=entity_type,
            positive=row["positive"],
            negative=row["negative"],
            neutral=row["neutral"],
            total=row["total"],
            avgScore=float(row["avg_score"]),
        )

    # ------------------------------------------------------------------
    # Helper
    # ------------------------------------------------------------------

    @staticmethod
    def _row_to_model(row: asyncpg.Record) -> SentimentResult:
        return SentimentResult(
            text=row["text"],
            label=SentimentLabel(row["label"]),
            score=row["score"],
            positiveWords=list(row["positive_words"]),
            negativeWords=list(row["negative_words"]),
            entityId=row["entity_id"],
            entityType=row["entity_type"],
            analyzedAt=row["analyzed_at"],
        )
