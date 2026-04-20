/**
 * Soak test — sustained moderate load for 2 hours to detect memory leaks,
 * connection pool exhaustion, and gradual performance degradation.
 * Run overnight; reduce to 30m for CI gate via K6_SOAK_DURATION env var.
 */
import http from 'k6/http';
import { group, sleep, check } from 'k6';
import { Trend, Rate, Counter, Gauge } from 'k6/metrics';
import {
  checkResponse, BASE_URL, DEFAULT_HEADERS,
  randomItem, sleep_random, errorRate,
} from '../lib/helpers.js';
import { PRODUCT_IDS, CATEGORY_IDS, SEARCH_QUERIES, TEST_USERS } from '../lib/data.js';

const soakErrorRate = new Rate('soak_error_rate');
const p95Trend = new Trend('soak_p95_window', true);
const memoryLeakIndicator = new Gauge('response_size_bytes');
const totalRequests = new Counter('soak_total_requests');

const DURATION = __ENV.K6_SOAK_DURATION || '2h';
const TARGET_VUS = parseInt(__ENV.K6_SOAK_VUS || '30');

export const options = {
  scenarios: {
    soak: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5m', target: TARGET_VUS },
        { duration: DURATION, target: TARGET_VUS },
        { duration: '5m', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<3000', 'p(99)<5000'],
    soak_error_rate: ['rate<0.01'],
    soak_p95_window: ['p(95)<3000'],
    error_rate: ['rate<0.01'],
  },
};

let requestIndex = 0;

export default function () {
  requestIndex++;
  totalRequests.add(1);

  // Rotate through all critical paths evenly
  const scenario = requestIndex % 6;

  switch (scenario) {
    case 0:
      group('Health', () => {
        const res = http.get(`${BASE_URL}/healthz`, { headers: DEFAULT_HEADERS });
        const ok = check(res, { 'healthz 200': (r) => r.status === 200 });
        soakErrorRate.add(!ok);
        p95Trend.add(res.timings.duration);
        sleep_random(200, 500);
      });
      break;

    case 1:
      group('Product List', () => {
        const page = (Math.floor(requestIndex / 6) % 10) + 1;
        const res = http.get(
          `${BASE_URL}/api/v1/products?page=${page}&size=20`,
          { headers: DEFAULT_HEADERS }
        );
        const ok = checkResponse(res, 'soak product list');
        soakErrorRate.add(!ok);
        p95Trend.add(res.timings.duration);
        if (res.body) memoryLeakIndicator.add(res.body.length);
        sleep_random(500, 1500);
      });
      break;

    case 2:
      group('Product Detail', () => {
        const productId = randomItem(PRODUCT_IDS);
        const res = http.get(
          `${BASE_URL}/api/v1/products/${productId}`,
          { headers: DEFAULT_HEADERS }
        );
        const ok = checkResponse(res, 'soak product detail');
        soakErrorRate.add(!ok);
        p95Trend.add(res.timings.duration);
        sleep_random(800, 2000);
      });
      break;

    case 3:
      group('Search', () => {
        const query = randomItem(SEARCH_QUERIES);
        const res = http.get(
          `${BASE_URL}/api/v1/search?q=${encodeURIComponent(query)}&page=1&size=10`,
          { headers: DEFAULT_HEADERS }
        );
        const ok = checkResponse(res, 'soak search');
        soakErrorRate.add(!ok);
        p95Trend.add(res.timings.duration);
        sleep_random(600, 1500);
      });
      break;

    case 4:
      group('Categories', () => {
        const res = http.get(`${BASE_URL}/api/v1/categories`, { headers: DEFAULT_HEADERS });
        const ok = checkResponse(res, 'soak categories');
        soakErrorRate.add(!ok);
        p95Trend.add(res.timings.duration);
        sleep_random(400, 1000);
      });
      break;

    case 5:
      group('Cart Create', () => {
        const user = randomItem(TEST_USERS);
        const res = http.post(
          `${BASE_URL}/api/v1/cart`,
          JSON.stringify({ userId: user.email }),
          { headers: DEFAULT_HEADERS }
        );
        const ok = checkResponse(res, 'soak cart create');
        soakErrorRate.add(!ok);
        p95Trend.add(res.timings.duration);
        sleep_random(400, 1000);
      });
      break;
  }

  sleep_random(200, 600);
}
