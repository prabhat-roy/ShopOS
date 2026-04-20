import random
from locust import TaskSet, task

PRODUCT_IDS = [
    "prod-001", "prod-002", "prod-003", "prod-004", "prod-005",
    "prod-010", "prod-020", "prod-030", "prod-040", "prod-050",
]

SHIPPING_ADDRESSES = [
    {"line1": "123 Test Street", "city": "New York", "state": "NY", "zip": "10001", "country": "US"},
    {"line1": "456 Load Ave", "city": "San Francisco", "state": "CA", "zip": "94105", "country": "US"},
]

PAYMENT_METHODS = [
    {"type": "card", "number": "4111111111111111", "expiry": "12/28", "cvv": "123"},
    {"type": "card", "number": "5500005555555559", "expiry": "11/27", "cvv": "456"},
]


def _auth_headers(token):
    return {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "Authorization": f"Bearer {token}",
    }


class CommerceTasks(TaskSet):
    def _get_token(self):
        return getattr(self.user, "token", None)

    def _get_cart_id(self):
        return getattr(self.user, "cart_id", None)

    @task(3)
    def create_cart(self):
        token = self._get_token()
        if not token:
            return
        with self.client.post(
            "/api/v1/cart",
            json={"userId": "load-test-user"},
            headers=_auth_headers(token),
            name="/api/v1/cart [POST]",
            catch_response=True,
        ) as resp:
            if resp.status_code in (200, 201):
                self.user.cart_id = resp.json().get("cartId")
                resp.success()
            else:
                resp.failure(f"Create cart: {resp.status_code}")

    @task(4)
    def add_to_cart(self):
        token = self._get_token()
        cart_id = self._get_cart_id()
        if not token or not cart_id:
            self.create_cart()
            return
        product_id = random.choice(PRODUCT_IDS)
        with self.client.post(
            f"/api/v1/cart/{cart_id}/items",
            json={"productId": product_id, "quantity": random.randint(1, 3)},
            headers=_auth_headers(token),
            name="/api/v1/cart/[id]/items [POST]",
            catch_response=True,
        ) as resp:
            if resp.status_code in (200, 201):
                resp.success()
            elif resp.status_code == 404:
                self.user.cart_id = None
                resp.failure("Cart not found — resetting")
            else:
                resp.failure(f"Add to cart: {resp.status_code}")

    @task(2)
    def view_cart(self):
        token = self._get_token()
        cart_id = self._get_cart_id()
        if not token or not cart_id:
            return
        self.client.get(
            f"/api/v1/cart/{cart_id}",
            headers=_auth_headers(token),
            name="/api/v1/cart/[id] [GET]",
        )

    @task(1)
    def checkout(self):
        token = self._get_token()
        cart_id = self._get_cart_id()
        if not token or not cart_id:
            return
        address = random.choice(SHIPPING_ADDRESSES)
        with self.client.post(
            "/api/v1/checkout",
            json={"cartId": cart_id, "shippingAddress": address, "shippingMethod": "standard"},
            headers=_auth_headers(token),
            name="/api/v1/checkout [POST]",
            catch_response=True,
        ) as resp:
            if resp.status_code in (200, 201):
                order_id = resp.json().get("orderId")
                if order_id:
                    self._pay(token, order_id)
                self.user.cart_id = None
                resp.success()
            else:
                resp.failure(f"Checkout: {resp.status_code}")

    def _pay(self, token, order_id):
        pm = random.choice(PAYMENT_METHODS)
        with self.client.post(
            "/api/v1/payments",
            json={"orderId": order_id, "paymentMethod": pm, "amount": {"value": 4999, "currency": "USD"}},
            headers=_auth_headers(token),
            name="/api/v1/payments [POST]",
            catch_response=True,
        ) as resp:
            if resp.status_code in (200, 201):
                resp.success()
            else:
                resp.failure(f"Payment: {resp.status_code}")
