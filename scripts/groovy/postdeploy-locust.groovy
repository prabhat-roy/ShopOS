def call() {
    def svc     = env.TEST_SERVICE
    def url     = env.SERVICE_URL
    def profile = env.LOAD_PROFILE ?: 'medium'

    sh 'mkdir -p reports/load/locust'

    def profiles = [
        light : [users: 10,  spawnRate: 2,  duration: '2m'],
        medium: [users: 50,  spawnRate: 5,  duration: '5m'],
        heavy : [users: 200, spawnRate: 20, duration: '10m'],
        spike : [users: 500, spawnRate: 50, duration: '90s'],
    ]
    def p         = profiles[profile] ?: profiles.medium
    def users     = env.LOAD_VUS?.trim() ? env.LOAD_VUS : p.users.toString()
    def spawnRate = p.spawnRate
    def duration  = env.LOAD_DURATION?.trim() ?: p.duration

    echo "Locust profile=${profile} users=${users} spawn-rate=${spawnRate} duration=${duration}"

    sh """
        mkdir -p /tmp/locust

        cat > /tmp/locust/locustfile.py << 'PYEOF'
from locust import HttpUser, task, between
import random
import uuid

PRODUCT_IDS = [f"prod-{i:04d}" for i in range(1, 101)]
USER_IDS    = [f"user-{i:04d}" for i in range(1, 51)]


class ShopOSUser(HttpUser):
    wait_time = between(1, 3)

    def on_start(self):
        self.user_id  = random.choice(USER_IDS)
        self.cart_id  = None
        self._login()

    def _login(self):
        resp = self.client.post("/api/v1/auth/login", json={
            "email":    f"{self.user_id}@test.shopos.io",
            "password": "TestPassword123!",
        }, name="/auth/login")
        if resp.status_code == 200:
            data = resp.json()
            self.client.headers["Authorization"] = f"Bearer {data.get('access_token', '')}"

    @task(5)
    def browse_products(self):
        self.client.get("/api/v1/products", params={"page": 1, "limit": 20},
                        name="/products (list)")

    @task(3)
    def search_products(self):
        queries = ["laptop", "phone", "headphones", "keyboard", "monitor"]
        self.client.get("/api/v1/products/search",
                        params={"q": random.choice(queries)},
                        name="/products/search")

    @task(2)
    def get_product(self):
        pid = random.choice(PRODUCT_IDS)
        self.client.get(f"/api/v1/products/{pid}", name="/products/{id}")

    @task(2)
    def add_to_cart(self):
        pid = random.choice(PRODUCT_IDS)
        resp = self.client.post("/api/v1/cart/items", json={
            "product_id": pid,
            "quantity": random.randint(1, 3),
        }, name="/cart/items (add)")
        if resp.status_code in (200, 201):
            self.cart_id = resp.json().get("cart_id")

    @task(1)
    def view_cart(self):
        self.client.get("/api/v1/cart", name="/cart (get)")

    @task(1)
    def checkout(self):
        if not self.cart_id:
            return
        self.client.post("/api/v1/checkout", json={
            "cart_id":            self.cart_id,
            "shipping_address_id": "addr-001",
            "payment_method_id":   "pm-test-visa",
        }, name="/checkout")
        self.cart_id = None

    @task(1)
    def view_orders(self):
        self.client.get("/api/v1/orders", name="/orders (list)")


class AdminUser(HttpUser):
    wait_time = between(2, 5)
    weight = 1

    def on_start(self):
        resp = self.client.post("/api/v1/auth/login", json={
            "email": "admin@shopos.io",
            "password": "AdminPassword123!",
        }, name="/auth/login (admin)")
        if resp.status_code == 200:
            token = resp.json().get("access_token", "")
            self.client.headers["Authorization"] = f"Bearer {token}"

    @task(3)
    def list_orders(self):
        self.client.get("/api/v1/admin/orders", params={"page": 1},
                        name="/admin/orders")

    @task(2)
    def get_analytics(self):
        self.client.get("/api/v1/admin/analytics/overview",
                        name="/admin/analytics")

    @task(1)
    def list_products(self):
        self.client.get("/api/v1/admin/products", params={"page": 1},
                        name="/admin/products")
PYEOF
    """

    sh """
        echo "=== Locust: ${users} users × ${duration} ==="
        docker run --rm \
            --network host \
            -v /tmp/locust:/mnt/locust \
            -v \${WORKSPACE}/reports/load/locust:/mnt/reports \
            locustio/locust:latest \
            --headless \
            --host ${url} \
            --users ${users} \
            --spawn-rate ${spawnRate} \
            --run-time ${duration} \
            --locustfile /mnt/locust/locustfile.py \
            --html /mnt/reports/locust-${svc}.html \
            --csv /mnt/reports/locust-${svc} \
            --csv-full-history \
            --logfile /mnt/reports/locust-${svc}.log \
            --exit-code-on-error 0 || true
        echo "Locust complete — reports/load/locust/"
    """

    if (profile == 'heavy' || profile == 'spike') {
        sh """
            echo "=== Locust distributed: master + 3 workers ==="
            docker run -d --name locust-master \
                --network host \
                -v /tmp/locust:/mnt/locust \
                -v \${WORKSPACE}/reports/load/locust:/mnt/reports \
                locustio/locust:latest \
                --master \
                --host ${url} \
                --users ${users} \
                --spawn-rate ${spawnRate} \
                --run-time ${duration} \
                --locustfile /mnt/locust/locustfile.py \
                --headless \
                --html /mnt/reports/locust-dist-${svc}.html \
                --csv /mnt/reports/locust-dist-${svc} \
                --expect-workers 3 || true

            sleep 3

            for i in 1 2 3; do
                docker run -d --name locust-worker-\$i \
                    --network host \
                    -v /tmp/locust:/mnt/locust \
                    locustio/locust:latest \
                    --worker \
                    --master-host localhost \
                    --locustfile /mnt/locust/locustfile.py || true
            done

            sleep \$(echo "${duration}" | sed 's/m/*60/' | bc 2>/dev/null || echo 300)

            docker stop locust-master locust-worker-1 locust-worker-2 locust-worker-3 2>/dev/null || true
            docker rm   locust-master locust-worker-1 locust-worker-2 locust-worker-3 2>/dev/null || true
        """
    }

    sh 'rm -rf /tmp/locust'
}
return this
