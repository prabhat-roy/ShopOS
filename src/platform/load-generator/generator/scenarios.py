import random
import time
from dataclasses import dataclass

import httpx


@dataclass
class ScenarioResult:
    scenario_name: str
    status_code: int
    latency_ms: float
    success: bool


async def browse_scenario(client: httpx.AsyncClient, base_url: str) -> ScenarioResult:
    """Simulate a user browsing the product catalogue."""
    start = time.perf_counter()
    last_status = 0
    success = True

    try:
        r1 = await client.get(f"{base_url}/products")
        last_status = r1.status_code
        if r1.status_code >= 400:
            success = False

        r2 = await client.get(f"{base_url}/products/random-id")
        last_status = r2.status_code
        if r2.status_code >= 400:
            success = False

        r3 = await client.get(f"{base_url}/categories")
        last_status = r3.status_code
        if r3.status_code >= 400:
            success = False
    except httpx.RequestError:
        last_status = 0
        success = False

    latency_ms = (time.perf_counter() - start) * 1000
    return ScenarioResult(
        scenario_name="browse",
        status_code=last_status,
        latency_ms=latency_ms,
        success=success,
    )


async def search_scenario(client: httpx.AsyncClient, base_url: str) -> ScenarioResult:
    """Simulate a user performing a search."""
    start = time.perf_counter()
    last_status = 0
    success = True

    try:
        r = await client.get(f"{base_url}/search", params={"q": "shoes"})
        last_status = r.status_code
        if r.status_code >= 400:
            success = False
    except httpx.RequestError:
        last_status = 0
        success = False

    latency_ms = (time.perf_counter() - start) * 1000
    return ScenarioResult(
        scenario_name="search",
        status_code=last_status,
        latency_ms=latency_ms,
        success=success,
    )


async def purchase_scenario(client: httpx.AsyncClient, base_url: str) -> ScenarioResult:
    """Simulate a user adding an item to cart and checking out."""
    start = time.perf_counter()
    last_status = 0
    success = True

    try:
        r1 = await client.post(
            f"{base_url}/cart/items",
            json={"product_id": "random-id", "quantity": 1},
        )
        last_status = r1.status_code
        if r1.status_code >= 400:
            success = False

        r2 = await client.get(f"{base_url}/cart")
        last_status = r2.status_code
        if r2.status_code >= 400:
            success = False

        r3 = await client.post(f"{base_url}/checkout", json={"cart_id": "random-cart-id"})
        last_status = r3.status_code
        if r3.status_code >= 400:
            success = False
    except httpx.RequestError:
        last_status = 0
        success = False

    latency_ms = (time.perf_counter() - start) * 1000
    return ScenarioResult(
        scenario_name="purchase",
        status_code=last_status,
        latency_ms=latency_ms,
        success=success,
    )


async def mixed_scenario(client: httpx.AsyncClient, base_url: str) -> ScenarioResult:
    """Randomly pick one of the three base scenarios."""
    scenario_fn = random.choice([browse_scenario, search_scenario, purchase_scenario])
    result = await scenario_fn(client, base_url)
    # Override the scenario name so callers know it was a mixed run.
    return ScenarioResult(
        scenario_name="mixed",
        status_code=result.status_code,
        latency_ms=result.latency_ms,
        success=result.success,
    )


SCENARIO_MAP = {
    "browse": browse_scenario,
    "search": search_scenario,
    "purchase": purchase_scenario,
    "mixed": mixed_scenario,
}
