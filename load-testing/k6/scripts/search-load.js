/**
 * Search load test — exercises the search-service under realistic query patterns:
 * keyword search → filtered search → autocomplete → zero-results
 */
import http from 'k6/http';
import { group, sleep, check } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import {
  checkResponse, BASE_URL, DEFAULT_HEADERS,
  randomItem, sleep_random, errorRate,
} from '../lib/helpers.js';
import { SEARCH_QUERIES, CATEGORY_IDS } from '../lib/data.js';

const searchLatency = new Trend('search_latency', true);
const autocompleteLatency = new Trend('autocomplete_latency', true);
const zeroResultsRate = new Rate('zero_results_rate');
const searchRequests = new Counter('search_requests_total');

export const options = {
  scenarios: {
    search: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 20 },
        { duration: '5m', target: 80 },
        { duration: '3m', target: 80 },
        { duration: '1m', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<2500', 'p(99)<4000'],
    search_latency: ['p(95)<2000', 'p(99)<3000'],
    autocomplete_latency: ['p(95)<300'],
    zero_results_rate: ['rate<0.10'],
    error_rate: ['rate<0.01'],
  },
};

const PRICE_RANGES = [
  { min: 0, max: 50 },
  { min: 50, max: 200 },
  { min: 200, max: 1000 },
];

const SORT_OPTIONS = ['relevance', 'price_asc', 'price_desc', 'newest', 'rating'];

export default function () {
  // ── Basic keyword search ────────────────────────────────────────────────────
  group('Keyword Search', () => {
    const query = randomItem(SEARCH_QUERIES);
    const t0 = Date.now();
    const res = http.get(
      `${BASE_URL}/api/v1/search?q=${encodeURIComponent(query)}&page=1&size=20`,
      { headers: DEFAULT_HEADERS }
    );
    searchRequests.add(1);
    const ok = checkResponse(res, 'keyword search');
    searchLatency.add(Date.now() - t0);

    if (ok && res.status === 200) {
      try {
        const body = JSON.parse(res.body);
        zeroResultsRate.add(body.total === 0);
      } catch (_) {}
    }
    sleep_random(800, 2000);
  });

  // ── Filtered search ─────────────────────────────────────────────────────────
  group('Filtered Search', () => {
    const query = randomItem(SEARCH_QUERIES);
    const cat = randomItem(CATEGORY_IDS);
    const price = randomItem(PRICE_RANGES);
    const sort = randomItem(SORT_OPTIONS);
    const t0 = Date.now();
    const res = http.get(
      `${BASE_URL}/api/v1/search?q=${encodeURIComponent(query)}&category=${cat}&price_min=${price.min}&price_max=${price.max}&sort=${sort}&page=1&size=20`,
      { headers: DEFAULT_HEADERS }
    );
    searchRequests.add(1);
    checkResponse(res, 'filtered search');
    searchLatency.add(Date.now() - t0);
    sleep_random(1000, 2500);
  });

  // ── Autocomplete ────────────────────────────────────────────────────────────
  group('Autocomplete', () => {
    const query = randomItem(SEARCH_QUERIES);
    // Simulate incremental typing — send partial queries
    for (let len = 2; len <= Math.min(query.length, 5); len++) {
      const partial = query.substring(0, len);
      const t0 = Date.now();
      const res = http.get(
        `${BASE_URL}/api/v1/search/suggest?q=${encodeURIComponent(partial)}&limit=5`,
        { headers: DEFAULT_HEADERS }
      );
      checkResponse(res, 'autocomplete');
      autocompleteLatency.add(Date.now() - t0);
      sleep_random(80, 150);
    }
    sleep_random(500, 1000);
  });

  // ── Zero-results query ──────────────────────────────────────────────────────
  group('Zero Results Query', () => {
    const nonsense = `zxqwerty-${Math.random().toString(36).substr(2, 8)}`;
    const t0 = Date.now();
    const res = http.get(
      `${BASE_URL}/api/v1/search?q=${nonsense}&page=1&size=10`,
      { headers: DEFAULT_HEADERS }
    );
    searchRequests.add(1);
    checkResponse(res, 'zero results search');
    searchLatency.add(Date.now() - t0);
    sleep_random(500, 1000);
  });

  // ── Facet aggregations ──────────────────────────────────────────────────────
  group('Facets', () => {
    const query = randomItem(SEARCH_QUERIES);
    const res = http.get(
      `${BASE_URL}/api/v1/search/facets?q=${encodeURIComponent(query)}`,
      { headers: DEFAULT_HEADERS }
    );
    checkResponse(res, 'search facets');
    sleep_random(600, 1500);
  });

  sleep_random(500, 1000);
}
