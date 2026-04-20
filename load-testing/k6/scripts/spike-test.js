/**
 * Spike test — sudden traffic surge to validate auto-scaling and circuit breakers.
 * Pattern: idle → instant spike to 500 VUs → hold briefly → drop back → recovery check
 */
import http from 'k6/http';
import { group, sleep, check } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import {
  checkResponse, BASE_URL, DEFAULT_HEADERS,
  randomItem, sleep_random, errorRate,
} from '../lib/helpers.js';
import { PRODUCT_IDS, SEARCH_QUERIES } from '../lib/data.js';

const spikeErrors = new Counter('spike_errors');
const recoveryRate = new Rate('recovery_rate');
const spikeDuration = new Trend('spike_request_duration', true);

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },   // baseline
        { duration: '30s', target: 500 },  // spike
        { duration: '1m',  target: 500 },  // hold spike
        { duration: '30s', target: 10 },   // drop
        { duration: '2m',  target: 10 },   // recovery validation
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    // Relaxed thresholds for spike — system must stay alive, not necessarily fast
    http_req_failed: ['rate<0.15'],
    http_req_duration: ['p(95)<10000'],
    spike_request_duration: ['p(99)<15000'],
    recovery_rate: ['rate>0.90'],
    error_rate: ['rate<0.15'],
  },
};

export default function () {
  const t0 = Date.now();

  // Spike focuses on the most cacheable read paths — heaviest fan-out
  group('Product List (Spike)', () => {
    const res = http.get(
      `${BASE_URL}/api/v1/products?page=1&size=20`,
      { headers: DEFAULT_HEADERS, timeout: '15s' }
    );
    const ok = checkResponse(res, 'product list spike');
    spikeDuration.add(Date.now() - t0);
    if (!ok) spikeErrors.add(1);
  });

  group('Product Detail (Spike)', () => {
    const productId = randomItem(PRODUCT_IDS);
    const res = http.get(
      `${BASE_URL}/api/v1/products/${productId}`,
      { headers: DEFAULT_HEADERS, timeout: '15s' }
    );
    checkResponse(res, 'product detail spike');
  });

  group('Search (Spike)', () => {
    const query = randomItem(SEARCH_QUERIES);
    const res = http.get(
      `${BASE_URL}/api/v1/search?q=${encodeURIComponent(query)}&page=1&size=10`,
      { headers: DEFAULT_HEADERS, timeout: '15s' }
    );
    const ok = checkResponse(res, 'search spike');
    // Track recovery: requests that succeed after the spike drops
    const elapsed = Date.now() - t0;
    if (elapsed > 150000) {  // after 2.5min = recovery phase
      recoveryRate.add(ok);
    }
  });

  group('Health Check (Spike)', () => {
    const res = http.get(`${BASE_URL}/healthz`, { headers: DEFAULT_HEADERS, timeout: '5s' });
    check(res, { 'healthz alive during spike': (r) => r.status === 200 });
  });

  sleep_random(100, 500);
}
