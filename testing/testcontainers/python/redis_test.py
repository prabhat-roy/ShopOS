"""
ShopOS — Redis Integration Tests with Testcontainers
======================================================
Spins up a real redis:7-alpine container and tests key/value operations,
TTL, pub/sub, and data structures used across ShopOS services.

Dependencies:
    pip install testcontainers[redis] redis pytest

Run:
    pytest testing/testcontainers/python/redis_test.py -v
"""

import json
import time
import threading
import uuid
from typing import Generator

import pytest
import redis as redis_lib
from testcontainers.redis import RedisContainer


# ── Fixtures ──────────────────────────────────────────────────────────────────

@pytest.fixture(scope="module")
def redis_container() -> Generator[RedisContainer, None, None]:
    """Start a redis:7-alpine container for the entire test module."""
    with RedisContainer(image="redis:7-alpine") as container:
        yield container


@pytest.fixture(scope="module")
def redis_client(redis_container: RedisContainer) -> Generator[redis_lib.Redis, None, None]:
    """Create a Redis client connected to the test container."""
    client = redis_lib.Redis(
        host=redis_container.get_container_host_ip(),
        port=redis_container.get_exposed_port(6379),
        db=0,
        decode_responses=True,
        socket_timeout=5,
        socket_connect_timeout=5,
    )
    yield client
    client.flushall()
    client.close()


@pytest.fixture(autouse=True)
def flush_between_tests(redis_client: redis_lib.Redis) -> Generator[None, None, None]:
    """Flush the DB between individual tests for isolation."""
    redis_client.flushdb()
    yield
    redis_client.flushdb()


# ── Basic String Tests ─────────────────────────────────────────────────────────

class TestRedisStringOperations:
    """Tests for basic string key/value — used by session-service, cart-service."""

    def test_set_and_get(self, redis_client: redis_lib.Redis) -> None:
        redis_client.set("greeting", "hello-shopos")
        value = redis_client.get("greeting")
        assert value == "hello-shopos"

    def test_get_nonexistent_key_returns_none(self, redis_client: redis_lib.Redis) -> None:
        value = redis_client.get("does-not-exist")
        assert value is None

    def test_set_with_ttl(self, redis_client: redis_lib.Redis) -> None:
        redis_client.setex("temp-key", 2, "expires-soon")
        assert redis_client.get("temp-key") == "expires-soon"
        ttl = redis_client.ttl("temp-key")
        assert 0 < ttl <= 2

    def test_key_expires_after_ttl(self, redis_client: redis_lib.Redis) -> None:
        redis_client.setex("short-lived", 1, "value")
        assert redis_client.get("short-lived") == "value"
        time.sleep(1.1)
        assert redis_client.get("short-lived") is None

    def test_set_nx_only_sets_if_absent(self, redis_client: redis_lib.Redis) -> None:
        set_result = redis_client.setnx("unique-key", "first")
        assert set_result is True
        second_result = redis_client.setnx("unique-key", "second")
        assert second_result is False
        assert redis_client.get("unique-key") == "first"

    def test_increment_counter(self, redis_client: redis_lib.Redis) -> None:
        redis_client.set("page:views", 0)
        for _ in range(5):
            redis_client.incr("page:views")
        assert redis_client.get("page:views") == "5"

    def test_delete_key(self, redis_client: redis_lib.Redis) -> None:
        redis_client.set("to-delete", "value")
        redis_client.delete("to-delete")
        assert redis_client.get("to-delete") is None


# ── Session Service Patterns ───────────────────────────────────────────────────

class TestSessionPatterns:
    """Simulates the session-service key patterns (session:{sessionId} → JSON)."""

    def test_store_and_retrieve_session(self, redis_client: redis_lib.Redis) -> None:
        session_id = str(uuid.uuid4())
        session_data = {
            "userId": str(uuid.uuid4()),
            "email": "user@shopos.dev",
            "roles": ["customer"],
            "deviceId": "device-abc123",
            "createdAt": "2026-04-23T10:00:00Z",
        }
        key = f"session:{session_id}"
        redis_client.setex(key, 3600, json.dumps(session_data))

        raw = redis_client.get(key)
        assert raw is not None
        retrieved = json.loads(raw)
        assert retrieved["userId"] == session_data["userId"]
        assert retrieved["email"] == "user@shopos.dev"
        assert "customer" in retrieved["roles"]

    def test_session_ttl_is_set(self, redis_client: redis_lib.Redis) -> None:
        session_id = str(uuid.uuid4())
        redis_client.setex(f"session:{session_id}", 3600, '{"userId":"u1"}')
        ttl = redis_client.ttl(f"session:{session_id}")
        assert 3500 < ttl <= 3600

    def test_delete_session_on_logout(self, redis_client: redis_lib.Redis) -> None:
        session_id = str(uuid.uuid4())
        redis_client.setex(f"session:{session_id}", 3600, '{"userId":"u1"}')
        redis_client.delete(f"session:{session_id}")
        assert redis_client.get(f"session:{session_id}") is None

    def test_multiple_sessions_for_same_user(self, redis_client: redis_lib.Redis) -> None:
        user_id = str(uuid.uuid4())
        session_ids = [str(uuid.uuid4()) for _ in range(3)]

        # Track user sessions in a set
        for sid in session_ids:
            redis_client.setex(f"session:{sid}", 3600, json.dumps({"userId": user_id}))
            redis_client.sadd(f"user:sessions:{user_id}", sid)

        members = redis_client.smembers(f"user:sessions:{user_id}")
        assert len(members) == 3
        for sid in session_ids:
            assert sid in members


# ── Cart Service Patterns ──────────────────────────────────────────────────────

class TestCartPatterns:
    """Simulates cart-service using Redis hashes (cart:{userId} → hash of items)."""

    def test_create_cart_as_hash(self, redis_client: redis_lib.Redis) -> None:
        user_id = str(uuid.uuid4())
        cart_key = f"cart:{user_id}"

        cart_data = {
            "currency": "USD",
            "coupon": "",
            "created_at": "2026-04-23T10:00:00Z",
        }
        redis_client.hset(cart_key, mapping=cart_data)
        redis_client.expire(cart_key, 259200)  # 72 hours

        assert redis_client.hget(cart_key, "currency") == "USD"
        assert redis_client.ttl(cart_key) > 0

    def test_add_items_to_cart(self, redis_client: redis_lib.Redis) -> None:
        user_id = str(uuid.uuid4())
        items_key = f"cart:{user_id}:items"

        items = [
            json.dumps({"productId": "prod-001", "variantId": "var-001", "quantity": 2, "price": 49.99}),
            json.dumps({"productId": "prod-002", "variantId": "var-002", "quantity": 1, "price": 99.99}),
        ]
        for item in items:
            redis_client.rpush(items_key, item)

        stored = redis_client.lrange(items_key, 0, -1)
        assert len(stored) == 2
        parsed = [json.loads(s) for s in stored]
        assert parsed[0]["productId"] == "prod-001"
        assert parsed[1]["quantity"] == 1


# ── Rate Limiter Patterns ──────────────────────────────────────────────────────

class TestRateLimiterPatterns:
    """Simulates rate-limiter-service using Redis atomic counters with TTL."""

    def test_sliding_window_rate_limit(self, redis_client: redis_lib.Redis) -> None:
        user_id = str(uuid.uuid4())
        window_key = f"ratelimit:{user_id}:api:v1"
        limit = 5
        window_seconds = 60

        for i in range(limit):
            count = redis_client.incr(window_key)
            if count == 1:
                redis_client.expire(window_key, window_seconds)
            assert count <= limit, f"Rate limit exceeded on request {i+1}"

        # Next request should exceed limit
        over_limit = redis_client.incr(window_key)
        assert over_limit > limit

    def test_rate_limit_resets_after_window(self, redis_client: redis_lib.Redis) -> None:
        user_id = str(uuid.uuid4())
        key = f"ratelimit:{user_id}:short"
        redis_client.setex(key, 1, 5)  # count=5, TTL=1s
        assert int(redis_client.get(key)) == 5

        time.sleep(1.1)
        assert redis_client.get(key) is None  # expired


# ── Compare / Recently Viewed Patterns ────────────────────────────────────────

class TestRecentlyViewedPatterns:
    """Simulates recently-viewed-service using Redis sorted sets (score = timestamp)."""

    def test_add_and_retrieve_recently_viewed(self, redis_client: redis_lib.Redis) -> None:
        user_id = str(uuid.uuid4())
        key = f"recently_viewed:{user_id}"

        products = ["prod-a", "prod-b", "prod-c", "prod-d", "prod-e"]
        for i, pid in enumerate(products):
            redis_client.zadd(key, {pid: time.time() + i})

        # Retrieve most recently viewed (highest score = most recent)
        recent = redis_client.zrevrange(key, 0, 4)
        assert len(recent) == 5
        assert recent[0] == "prod-e"  # last added → highest score

    def test_cap_recently_viewed_at_20(self, redis_client: redis_lib.Redis) -> None:
        user_id = str(uuid.uuid4())
        key = f"recently_viewed:{user_id}"

        for i in range(25):
            redis_client.zadd(key, {f"prod-{i:03d}": float(i)})
            # Trim to keep only 20 most recent
            redis_client.zremrangebyrank(key, 0, -21)

        count = redis_client.zcard(key)
        assert count == 20


# ── Pub/Sub Tests ─────────────────────────────────────────────────────────────

class TestPubSub:
    """Tests Redis pub/sub — used by in-app-notification-service and live-chat-service."""

    def test_publish_and_subscribe(self, redis_container: RedisContainer) -> None:
        publisher = redis_lib.Redis(
            host=redis_container.get_container_host_ip(),
            port=redis_container.get_exposed_port(6379),
            db=0,
            decode_responses=True,
        )
        subscriber = redis_lib.Redis(
            host=redis_container.get_container_host_ip(),
            port=redis_container.get_exposed_port(6379),
            db=0,
            decode_responses=True,
        )

        channel = f"notifications:user:{uuid.uuid4()}"
        received_messages = []

        pubsub = subscriber.pubsub()
        pubsub.subscribe(channel)

        def listen():
            for msg in pubsub.listen():
                if msg["type"] == "message":
                    received_messages.append(msg["data"])
                    break

        listener_thread = threading.Thread(target=listen, daemon=True)
        listener_thread.start()

        # Give subscriber time to register
        time.sleep(0.1)

        notification_payload = json.dumps({
            "type": "order.confirmed",
            "orderId": str(uuid.uuid4()),
            "message": "Your order has been confirmed!",
        })
        publisher.publish(channel, notification_payload)

        listener_thread.join(timeout=5)
        pubsub.unsubscribe(channel)

        assert len(received_messages) == 1
        parsed = json.loads(received_messages[0])
        assert parsed["type"] == "order.confirmed"

        publisher.close()
        subscriber.close()


# ── Server Info ───────────────────────────────────────────────────────────────

class TestRedisServerInfo:
    """Sanity checks on the Redis server version and configuration."""

    def test_redis_version_is_7(self, redis_client: redis_lib.Redis) -> None:
        info = redis_client.info("server")
        version = info.get("redis_version", "")
        major = int(version.split(".")[0]) if version else 0
        assert major >= 7, f"Expected Redis 7+, got {version}"

    def test_ping(self, redis_client: redis_lib.Redis) -> None:
        assert redis_client.ping() is True

    def test_flushdb_clears_data(self, redis_client: redis_lib.Redis) -> None:
        redis_client.set("key1", "val1")
        redis_client.set("key2", "val2")
        redis_client.flushdb()
        assert redis_client.dbsize() == 0
