import random
import urllib.parse
from locust import TaskSet, task

SEARCH_QUERIES = [
    "laptop", "smartphone", "headphones", "keyboard", "monitor",
    "shirt", "shoes", "jeans", "jacket", "dress",
    "python book", "coffee maker", "yoga mat", "running shoes",
]

SORT_OPTIONS = ["relevance", "price_asc", "price_desc", "newest", "rating"]


class SearchTasks(TaskSet):
    @task(5)
    def keyword_search(self):
        query = random.choice(SEARCH_QUERIES)
        page = random.randint(1, 5)
        encoded = urllib.parse.quote(query)
        self.client.get(
            f"/api/v1/search?q={encoded}&page={page}&size=20",
            name="/api/v1/search",
        )

    @task(3)
    def filtered_search(self):
        query = random.choice(SEARCH_QUERIES)
        sort = random.choice(SORT_OPTIONS)
        price_max = random.choice([50, 200, 500, 1000])
        encoded = urllib.parse.quote(query)
        self.client.get(
            f"/api/v1/search?q={encoded}&price_max={price_max}&sort={sort}&page=1&size=20",
            name="/api/v1/search?filtered",
        )

    @task(4)
    def autocomplete(self):
        query = random.choice(SEARCH_QUERIES)
        # Partial query (2–5 chars)
        partial = query[: random.randint(2, min(5, len(query)))]
        encoded = urllib.parse.quote(partial)
        self.client.get(
            f"/api/v1/search/suggest?q={encoded}&limit=5",
            name="/api/v1/search/suggest",
        )

    @task(1)
    def search_facets(self):
        query = random.choice(SEARCH_QUERIES)
        encoded = urllib.parse.quote(query)
        self.client.get(
            f"/api/v1/search/facets?q={encoded}",
            name="/api/v1/search/facets",
        )

    @task(1)
    def zero_results(self):
        """Ensures system handles zero-result gracefully."""
        nonsense = f"zxqwerty-{random.randint(10000, 99999)}"
        self.client.get(
            f"/api/v1/search?q={nonsense}&page=1&size=10",
            name="/api/v1/search?zero-results",
        )
