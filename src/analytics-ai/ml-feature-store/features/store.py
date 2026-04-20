from __future__ import annotations

import json
from datetime import datetime, timezone
from typing import Any, Optional

import asyncpg

from .models import FeatureDefinition, FeatureType, FeatureValue, FeatureVector


CREATE_DEFINITIONS_TABLE = """
CREATE TABLE IF NOT EXISTS feature_definitions (
    name          TEXT        NOT NULL,
    feature_group TEXT        NOT NULL,
    type          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    tags          JSONB       NOT NULL DEFAULT '[]',
    default_value JSONB,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY   (name, feature_group)
);
"""

CREATE_VALUES_TABLE = """
CREATE TABLE IF NOT EXISTS feature_values (
    id            BIGSERIAL PRIMARY KEY,
    entity_id     TEXT        NOT NULL,
    feature_name  TEXT        NOT NULL,
    feature_group TEXT        NOT NULL,
    value         JSONB       NOT NULL,
    version       INTEGER     NOT NULL DEFAULT 1,
    computed_at   TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_values_unique
    ON feature_values (entity_id, feature_name, feature_group);
CREATE INDEX IF NOT EXISTS idx_feature_values_entity_group
    ON feature_values (entity_id, feature_group);
"""


class AsyncPgStore:
    def __init__(self, database_url: str) -> None:
        self._database_url = database_url
        self._pool: Optional[asyncpg.Pool] = None

    async def connect(self) -> None:
        self._pool = await asyncpg.create_pool(self._database_url, min_size=2, max_size=10)
        async with self._pool.acquire() as conn:
            await conn.execute(CREATE_DEFINITIONS_TABLE)
            await conn.execute(CREATE_VALUES_TABLE)

    async def disconnect(self) -> None:
        if self._pool:
            await self._pool.close()

    # ------------------------------------------------------------------
    # Feature Definitions
    # ------------------------------------------------------------------

    async def register_feature(self, definition: FeatureDefinition) -> FeatureDefinition:
        now = datetime.now(timezone.utc)
        sql = """
            INSERT INTO feature_definitions
                (name, feature_group, type, description, tags, default_value, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8)
            ON CONFLICT (name, feature_group)
            DO UPDATE SET
                type          = EXCLUDED.type,
                description   = EXCLUDED.description,
                tags          = EXCLUDED.tags,
                default_value = EXCLUDED.default_value,
                updated_at    = EXCLUDED.updated_at
            RETURNING name, feature_group, type, description, tags, default_value
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                sql,
                definition.name,
                definition.featureGroup,
                definition.type.value,
                definition.description,
                json.dumps(definition.tags),
                json.dumps(definition.defaultValue),
                now,
                now,
            )
        return _row_to_definition(row)

    async def get_definition(
        self, name: str, group: str
    ) -> Optional[FeatureDefinition]:
        sql = """
            SELECT name, feature_group, type, description, tags, default_value
            FROM   feature_definitions
            WHERE  name = $1 AND feature_group = $2
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(sql, name, group)
        if row is None:
            return None
        return _row_to_definition(row)

    async def list_definitions(
        self,
        group: Optional[str] = None,
        feature_type: Optional[FeatureType] = None,
    ) -> list[FeatureDefinition]:
        conditions: list[str] = []
        params: list[Any] = []
        if group is not None:
            params.append(group)
            conditions.append(f"feature_group = ${len(params)}")
        if feature_type is not None:
            params.append(feature_type.value)
            conditions.append(f"type = ${len(params)}")

        where = f"WHERE {' AND '.join(conditions)}" if conditions else ""
        sql = f"""
            SELECT name, feature_group, type, description, tags, default_value
            FROM   feature_definitions
            {where}
            ORDER  BY feature_group, name
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql, *params)
        return [_row_to_definition(r) for r in rows]

    # ------------------------------------------------------------------
    # Feature Values
    # ------------------------------------------------------------------

    async def save_value(self, fv: FeatureValue) -> None:
        sql = """
            INSERT INTO feature_values
                (entity_id, feature_name, feature_group, value, version, computed_at)
            VALUES ($1, $2, $3, $4::jsonb, $5, $6)
            ON CONFLICT (entity_id, feature_name, feature_group)
            DO UPDATE SET
                value       = EXCLUDED.value,
                version     = EXCLUDED.version,
                computed_at = EXCLUDED.computed_at
        """
        async with self._pool.acquire() as conn:
            await conn.execute(
                sql,
                fv.entityId,
                fv.featureName,
                fv.featureGroup,
                json.dumps(fv.value),
                fv.version,
                fv.computedAt,
            )

    async def get_value(
        self, entity_id: str, feature_name: str, group: str
    ) -> Optional[FeatureValue]:
        sql = """
            SELECT entity_id, feature_name, feature_group, value, version, computed_at
            FROM   feature_values
            WHERE  entity_id = $1 AND feature_name = $2 AND feature_group = $3
        """
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(sql, entity_id, feature_name, group)
        if row is None:
            return None
        return _row_to_value(row)

    async def save_batch(self, values: list[FeatureValue]) -> None:
        for fv in values:
            await self.save_value(fv)

    async def get_feature_vector(
        self, entity_id: str, feature_names: list[str], group: str
    ) -> FeatureVector:
        sql = """
            SELECT feature_name, value
            FROM   feature_values
            WHERE  entity_id = $1 AND feature_group = $2
              AND  feature_name = ANY($3)
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql, entity_id, group, feature_names)

        found: dict[str, Any] = {}
        for row in rows:
            raw = row["value"]
            found[row["feature_name"]] = json.loads(raw) if isinstance(raw, str) else raw

        missing: list[str] = []
        features: dict[str, Any] = {}

        for name in feature_names:
            if name in found:
                features[name] = found[name]
            else:
                # Try to fill from definition default
                definition = await self.get_definition(name, group)
                if definition is not None and definition.defaultValue is not None:
                    features[name] = definition.defaultValue
                else:
                    missing.append(name)

        return FeatureVector(
            entityId=entity_id,
            features=features,
            missingFeatures=missing,
            retrievedAt=datetime.now(timezone.utc),
        )

    async def get_entity_features(
        self, entity_id: str, group: str
    ) -> list[FeatureValue]:
        sql = """
            SELECT entity_id, feature_name, feature_group, value, version, computed_at
            FROM   feature_values
            WHERE  entity_id = $1 AND feature_group = $2
            ORDER  BY feature_name
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql, entity_id, group)
        return [_row_to_value(r) for r in rows]

    async def delete_entity_features(self, entity_id: str, group: str) -> int:
        sql = """
            DELETE FROM feature_values
            WHERE entity_id = $1 AND feature_group = $2
        """
        async with self._pool.acquire() as conn:
            result = await conn.execute(sql, entity_id, group)
        # asyncpg returns "DELETE N"
        try:
            return int(result.split()[-1])
        except (IndexError, ValueError):
            return 0


def _row_to_definition(row: asyncpg.Record) -> FeatureDefinition:
    tags_raw = row["tags"]
    tags = json.loads(tags_raw) if isinstance(tags_raw, str) else tags_raw

    default_raw = row["default_value"]
    if default_raw is None:
        default_value = None
    elif isinstance(default_raw, str):
        default_value = json.loads(default_raw)
    else:
        default_value = default_raw

    return FeatureDefinition(
        name=row["name"],
        featureGroup=row["feature_group"],
        type=FeatureType(row["type"]),
        description=row["description"],
        tags=tags if tags else [],
        defaultValue=default_value,
    )


def _row_to_value(row: asyncpg.Record) -> FeatureValue:
    raw = row["value"]
    value = json.loads(raw) if isinstance(raw, str) else raw
    return FeatureValue(
        entityId=row["entity_id"],
        featureName=row["feature_name"],
        featureGroup=row["feature_group"],
        value=value,
        version=row["version"],
        computedAt=row["computed_at"],
    )
