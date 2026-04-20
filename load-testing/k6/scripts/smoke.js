/**
 * Smoke test — 1 VU, 2 minutes
 * Verifies all critical paths return 200 and respond within 2s.
 */
import http from 'k6/http';
import { group, sleep } from 'k6';
import { checkResponse, BASE_URL, DEFAULT_HEADERS } from '../lib/helpers.js';

export const options = {
  vus: 1,
  duration: '2m',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<2000'],
    error_rate: ['rate<0.01'],
  },
};

export default function () {
  group('Health Checks', () => {
    const res = http.get(`${BASE_URL}/healthz`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'api-gateway healthz');
    sleep(0.5);
  });

  group('Product Catalog', () => {
    let res = http.get(`${BASE_URL}/api/v1/products?page=1&size=10`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'list products');
    sleep(0.3);

    res = http.get(`${BASE_URL}/api/v1/products/prod-001`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'get product');
    sleep(0.3);

    res = http.get(`${BASE_URL}/api/v1/categories`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'list categories');
    sleep(0.3);
  });

  group('Search', () => {
    const res = http.get(`${BASE_URL}/api/v1/search?q=laptop&page=1&size=10`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'search products');
    sleep(0.5);
  });

  group('Cart', () => {
    const cartRes = http.post(`${BASE_URL}/api/v1/cart`, JSON.stringify({
      userId: 'smoke-user-001',
    }), { headers: DEFAULT_HEADERS });
    checkResponse(cartRes, 'create cart');

    if (cartRes.status === 201 || cartRes.status === 200) {
      const cartId = JSON.parse(cartRes.body).cartId;
      const addRes = http.post(`${BASE_URL}/api/v1/cart/${cartId}/items`, JSON.stringify({
        productId: 'prod-001',
        quantity: 1,
      }), { headers: DEFAULT_HEADERS });
      checkResponse(addRes, 'add to cart');
    }
    sleep(0.5);
  });

  sleep(1);
}
