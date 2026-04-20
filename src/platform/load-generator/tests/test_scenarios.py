import pytest
import respx
import httpx
from httpx import Response

from generator.scenarios import (
    ScenarioResult,
    browse_scenario,
    mixed_scenario,
    purchase_scenario,
    search_scenario,
)

BASE_URL = "http://target"


# ---------------------------------------------------------------------------
# browse_scenario
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_browse_scenario_success():
    respx.get(f"{BASE_URL}/products").mock(return_value=Response(200, json=[]))
    respx.get(f"{BASE_URL}/products/random-id").mock(return_value=Response(200, json={}))
    respx.get(f"{BASE_URL}/categories").mock(return_value=Response(200, json=[]))

    async with httpx.AsyncClient() as client:
        result = await browse_scenario(client, BASE_URL)

    assert isinstance(result, ScenarioResult)
    assert result.scenario_name == "browse"
    assert result.status_code == 200
    assert result.latency_ms >= 0
    assert result.success is True


@pytest.mark.asyncio
@respx.mock
async def test_browse_scenario_failure():
    respx.get(f"{BASE_URL}/products").mock(return_value=Response(500))
    respx.get(f"{BASE_URL}/products/random-id").mock(return_value=Response(500))
    respx.get(f"{BASE_URL}/categories").mock(return_value=Response(500))

    async with httpx.AsyncClient() as client:
        result = await browse_scenario(client, BASE_URL)

    assert result.scenario_name == "browse"
    assert result.success is False
    assert result.status_code == 500


# ---------------------------------------------------------------------------
# search_scenario
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_search_scenario_success():
    respx.get(f"{BASE_URL}/search").mock(return_value=Response(200, json={"hits": []}))

    async with httpx.AsyncClient() as client:
        result = await search_scenario(client, BASE_URL)

    assert isinstance(result, ScenarioResult)
    assert result.scenario_name == "search"
    assert result.status_code == 200
    assert result.latency_ms >= 0
    assert result.success is True


@pytest.mark.asyncio
@respx.mock
async def test_search_scenario_failure():
    respx.get(f"{BASE_URL}/search").mock(return_value=Response(404))

    async with httpx.AsyncClient() as client:
        result = await search_scenario(client, BASE_URL)

    assert result.scenario_name == "search"
    assert result.success is False


# ---------------------------------------------------------------------------
# purchase_scenario
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_purchase_scenario_success():
    respx.post(f"{BASE_URL}/cart/items").mock(return_value=Response(201, json={}))
    respx.get(f"{BASE_URL}/cart").mock(return_value=Response(200, json={}))
    respx.post(f"{BASE_URL}/checkout").mock(return_value=Response(200, json={}))

    async with httpx.AsyncClient() as client:
        result = await purchase_scenario(client, BASE_URL)

    assert isinstance(result, ScenarioResult)
    assert result.scenario_name == "purchase"
    assert result.success is True
    assert result.latency_ms >= 0


@pytest.mark.asyncio
@respx.mock
async def test_purchase_scenario_failure():
    respx.post(f"{BASE_URL}/cart/items").mock(return_value=Response(400))
    respx.get(f"{BASE_URL}/cart").mock(return_value=Response(400))
    respx.post(f"{BASE_URL}/checkout").mock(return_value=Response(400))

    async with httpx.AsyncClient() as client:
        result = await purchase_scenario(client, BASE_URL)

    assert result.scenario_name == "purchase"
    assert result.success is False


# ---------------------------------------------------------------------------
# mixed_scenario
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_mixed_scenario_returns_mixed_name():
    # Stub all endpoints that any sub-scenario might call.
    respx.get(f"{BASE_URL}/products").mock(return_value=Response(200, json=[]))
    respx.get(f"{BASE_URL}/products/random-id").mock(return_value=Response(200, json={}))
    respx.get(f"{BASE_URL}/categories").mock(return_value=Response(200, json=[]))
    respx.get(f"{BASE_URL}/search").mock(return_value=Response(200, json={}))
    respx.post(f"{BASE_URL}/cart/items").mock(return_value=Response(201, json={}))
    respx.get(f"{BASE_URL}/cart").mock(return_value=Response(200, json={}))
    respx.post(f"{BASE_URL}/checkout").mock(return_value=Response(200, json={}))

    async with httpx.AsyncClient() as client:
        result = await mixed_scenario(client, BASE_URL)

    assert isinstance(result, ScenarioResult)
    assert result.scenario_name == "mixed"
    assert result.latency_ms >= 0
    assert isinstance(result.success, bool)
