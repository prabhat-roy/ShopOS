"""
ShopOS Locust entry point.
Registers all task sets and exposes composite users for different load profiles.
"""
import os
from locust import HttpUser, between, events
from locust.contrib.fasthttp import FastHttpUser

from tasks.auth import AuthTasks
from tasks.catalog import CatalogTasks
from tasks.commerce import CommerceTasks
from tasks.search import SearchTasks

BASE_URL = os.getenv("BASE_URL", "http://api-gateway.platform.svc.cluster.local:8080")


class BrowseUser(FastHttpUser):
    """Anonymous catalog browser — highest volume, read-only."""
    host = BASE_URL
    wait_time = between(0.5, 2.0)
    tasks = {
        CatalogTasks: 7,
        SearchTasks: 3,
    }


class ShopperUser(HttpUser):
    """Authenticated shopper — browses, searches, adds to cart, checks out."""
    host = BASE_URL
    wait_time = between(1.0, 3.0)
    tasks = {
        AuthTasks: 1,
        CatalogTasks: 4,
        SearchTasks: 3,
        CommerceTasks: 2,
    }

    def on_start(self):
        """Login before executing tasks."""
        self.token = None
        self.cart_id = None
        response = self.client.post(
            "/api/v1/auth/login",
            json={"email": "user1@test.shopos.local", "password": "Password1!"},
        )
        if response.status_code == 200:
            self.token = response.json().get("access_token")

    def on_stop(self):
        self.token = None
        self.cart_id = None


class PowerUser(HttpUser):
    """Heavy user — full purchase flow including payment."""
    host = BASE_URL
    wait_time = between(0.5, 1.5)
    tasks = {
        CommerceTasks: 6,
        CatalogTasks: 3,
        SearchTasks: 1,
    }

    def on_start(self):
        self.token = None
        self.cart_id = None
        response = self.client.post(
            "/api/v1/auth/login",
            json={"email": "user3@test.shopos.local", "password": "Password3!"},
        )
        if response.status_code == 200:
            self.token = response.json().get("access_token")


@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    print(f"[ShopOS] Starting load test against {BASE_URL}")


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    print("[ShopOS] Load test complete")
