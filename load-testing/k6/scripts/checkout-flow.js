/**
 * Checkout flow load test — simulates the full purchase journey:
 * browse → search → product detail → add to cart → checkout → payment
 */
import http from 'k6/http';
import { group, sleep, check } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import {
  checkResponse, BASE_URL, DEFAULT_HEADERS,
  randomItem, sleep_random, errorRate,
} from '../lib/helpers.js';
import { PRODUCT_IDS, SEARCH_QUERIES, TEST_USERS, PAYMENT_METHODS, SHIPPING_ADDRESSES } from '../lib/data.js';

const checkoutDuration = new Trend('checkout_duration', true);
const orderCreated = new Counter('orders_created');
const paymentSuccessRate = new Rate('payment_success_rate');

export const options = {
  scenarios: {
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 20 },
        { duration: '5m', target: 50 },
        { duration: '2m', target: 50 },
        { duration: '1m', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<3000', 'p(99)<5000'],
    checkout_duration: ['p(95)<5000'],
    payment_success_rate: ['rate>0.95'],
    error_rate: ['rate<0.05'],
  },
};

export default function () {
  const user = randomItem(TEST_USERS);
  let token = null;

  // ── Step 1: Authenticate ────────────────────────────────────────────────────
  group('Authentication', () => {
    const res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
      email: user.email,
      password: user.password,
    }), { headers: DEFAULT_HEADERS });

    if (check(res, { 'login 200': (r) => r.status === 200 })) {
      token = JSON.parse(res.body).access_token;
    }
    sleep_random(300, 800);
  });

  if (!token) return;

  const headers = { ...DEFAULT_HEADERS, Authorization: `Bearer ${token}` };

  // ── Step 2: Browse products ─────────────────────────────────────────────────
  group('Browse Products', () => {
    const res = http.get(
      `${BASE_URL}/api/v1/products?page=1&size=20`,
      { headers }
    );
    checkResponse(res, 'browse products');
    sleep_random(500, 2000);
  });

  // ── Step 3: Search ──────────────────────────────────────────────────────────
  group('Search', () => {
    const query = randomItem(SEARCH_QUERIES);
    const res = http.get(
      `${BASE_URL}/api/v1/search?q=${encodeURIComponent(query)}&page=1&size=10`,
      { headers }
    );
    checkResponse(res, 'search');
    sleep_random(800, 2000);
  });

  // ── Step 4: Product detail ──────────────────────────────────────────────────
  const productId = randomItem(PRODUCT_IDS);
  group('Product Detail', () => {
    const res = http.get(`${BASE_URL}/api/v1/products/${productId}`, { headers });
    checkResponse(res, 'product detail');
    sleep_random(1000, 3000);
  });

  // ── Step 5: Add to cart ─────────────────────────────────────────────────────
  let cartId = null;
  group('Add to Cart', () => {
    const cartRes = http.post(`${BASE_URL}/api/v1/cart`, JSON.stringify({
      userId: user.email,
    }), { headers });
    checkResponse(cartRes, 'create cart');

    if (cartRes.status === 200 || cartRes.status === 201) {
      cartId = JSON.parse(cartRes.body).cartId;
      const addRes = http.post(`${BASE_URL}/api/v1/cart/${cartId}/items`, JSON.stringify({
        productId,
        quantity: 1,
      }), { headers });
      checkResponse(addRes, 'add to cart');
    }
    sleep_random(500, 1500);
  });

  if (!cartId) return;

  // ── Step 6: Checkout ────────────────────────────────────────────────────────
  let orderId = null;
  const checkoutStart = Date.now();
  group('Checkout', () => {
    const address = randomItem(SHIPPING_ADDRESSES);
    const res = http.post(`${BASE_URL}/api/v1/checkout`, JSON.stringify({
      cartId,
      shippingAddress: address,
      shippingMethod: 'standard',
    }), { headers });
    checkResponse(res, 'checkout');

    if (res.status === 200 || res.status === 201) {
      orderId = JSON.parse(res.body).orderId;
    }
    sleep_random(300, 800);
  });

  // ── Step 7: Payment ─────────────────────────────────────────────────────────
  if (orderId) {
    group('Payment', () => {
      const pm = randomItem(PAYMENT_METHODS);
      const res = http.post(`${BASE_URL}/api/v1/payments`, JSON.stringify({
        orderId,
        paymentMethod: pm,
        amount: { value: 4999, currency: 'USD' },
      }), { headers });

      const paid = check(res, { 'payment success': (r) => r.status === 200 || r.status === 201 });
      paymentSuccessRate.add(paid);
      if (paid) {
        orderCreated.add(1);
        checkoutDuration.add(Date.now() - checkoutStart);
      }
    });
  }

  sleep_random(1000, 3000);
}
