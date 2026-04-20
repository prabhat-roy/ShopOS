from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from main import app


@pytest.fixture
def client():
    app.state.redis = None
    with TestClient(app) as c:
        yield c


def test_healthz(client):
    res = client.get("/healthz")
    assert res.status_code == 200
    assert res.json() == {"status": "ok"}


def test_chat_greeting(client):
    res = client.post("/chat", json={"session_id": "s1", "message": "hello"})
    assert res.status_code == 200
    data = res.json()
    assert data["intent"] == "greeting"
    assert data["escalate"] is False
    assert "session_id" in data
    assert len(data["reply"]) > 0


def test_chat_order_status(client):
    res = client.post("/chat", json={"session_id": "s2", "message": "where is my order?"})
    assert res.status_code == 200
    assert res.json()["intent"] == "order_status"


def test_chat_return(client):
    res = client.post("/chat", json={"session_id": "s3", "message": "I want to return my item"})
    assert res.status_code == 200
    assert res.json()["intent"] == "return_refund"


def test_chat_unknown(client):
    res = client.post("/chat", json={"session_id": "s4", "message": "abcdefxyz random text"})
    assert res.status_code == 200
    assert res.json()["intent"] == "unknown"
