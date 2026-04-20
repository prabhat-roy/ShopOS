import random
from locust import TaskSet, task

PRODUCT_IDS = [
    "prod-001", "prod-002", "prod-003", "prod-004", "prod-005",
    "prod-010", "prod-020", "prod-030", "prod-040", "prod-050",
    "prod-100", "prod-200", "prod-300", "prod-400", "prod-500",
]

CATEGORY_IDS = [
    "cat-electronics", "cat-clothing", "cat-books",
    "cat-home", "cat-sports", "cat-beauty",
]


class CatalogTasks(TaskSet):
    @task(5)
    def list_products(self):
        page = random.randint(1, 10)
        self.client.get(
            f"/api/v1/products?page={page}&size=24&sort=relevance",
            name="/api/v1/products",
        )

    @task(4)
    def product_detail(self):
        product_id = random.choice(PRODUCT_IDS)
        self.client.get(
            f"/api/v1/products/{product_id}",
            name="/api/v1/products/[id]",
        )

    @task(3)
    def list_categories(self):
        self.client.get("/api/v1/categories", name="/api/v1/categories")

    @task(3)
    def category_products(self):
        cat = random.choice(CATEGORY_IDS)
        page = random.randint(1, 5)
        self.client.get(
            f"/api/v1/products?category={cat}&page={page}&size=24",
            name="/api/v1/products?category=[id]",
        )

    @task(2)
    def related_products(self):
        product_id = random.choice(PRODUCT_IDS)
        self.client.get(
            f"/api/v1/products/{product_id}/related?limit=8",
            name="/api/v1/products/[id]/related",
        )

    @task(1)
    def featured_products(self):
        self.client.get(
            "/api/v1/products/featured?limit=12",
            name="/api/v1/products/featured",
        )

    @task(1)
    def healthcheck(self):
        with self.client.get(
            "/healthz",
            name="/healthz",
            catch_response=True,
        ) as resp:
            if resp.status_code != 200:
                resp.failure(f"Health check returned {resp.status_code}")
