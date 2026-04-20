/**
 * Product browse load test — simulates catalog browsing behavior:
 * homepage → categories → product listing → product detail → related products
 */
import http from 'k6/http';
import { group, sleep, check } from 'k6';
import { Trend, Rate } from 'k6/metrics';
import {
  checkResponse, BASE_URL, DEFAULT_HEADERS,
  randomItem, sleep_random, errorRate,
} from '../lib/helpers.js';
import { PRODUCT_IDS, CATEGORY_IDS, SEARCH_QUERIES } from '../lib/data.js';

const catalogPageDuration = new Trend('catalog_page_duration', true);
const productDetailDuration = new Trend('product_detail_duration', true);
const cacheHitRate = new Rate('cache_hit_rate');

export const options = {
  scenarios: {
    browse: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 30 },
        { duration: '5m', target: 100 },
        { duration: '3m', target: 100 },
        { duration: '1m', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<2000', 'p(99)<3000'],
    catalog_page_duration: ['p(95)<1500'],
    product_detail_duration: ['p(95)<1000'],
    error_rate: ['rate<0.01'],
  },
};

export default function () {
  // ── Homepage / Featured Products ───────────────────────────────────────────
  group('Homepage', () => {
    const t0 = Date.now();
    const res = http.get(`${BASE_URL}/api/v1/products/featured?limit=12`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'featured products');
    catalogPageDuration.add(Date.now() - t0);

    const isCached = res.headers['X-Cache'] === 'HIT' || res.headers['x-cache'] === 'HIT';
    cacheHitRate.add(isCached);
    sleep_random(500, 1500);
  });

  // ── Category Listing ────────────────────────────────────────────────────────
  group('Category Listing', () => {
    const t0 = Date.now();
    const res = http.get(`${BASE_URL}/api/v1/categories`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'categories');
    catalogPageDuration.add(Date.now() - t0);
    sleep_random(300, 800);
  });

  // ── Category Products ───────────────────────────────────────────────────────
  group('Category Products', () => {
    const catId = randomItem(CATEGORY_IDS);
    const t0 = Date.now();
    const res = http.get(
      `${BASE_URL}/api/v1/products?category=${catId}&page=1&size=24&sort=relevance`,
      { headers: DEFAULT_HEADERS }
    );
    checkResponse(res, 'category products');
    catalogPageDuration.add(Date.now() - t0);

    const isCached = res.headers['X-Cache'] === 'HIT' || res.headers['x-cache'] === 'HIT';
    cacheHitRate.add(isCached);
    sleep_random(1000, 3000);
  });

  // ── Product Detail ──────────────────────────────────────────────────────────
  group('Product Detail', () => {
    const productId = randomItem(PRODUCT_IDS);
    const t0 = Date.now();

    const res = http.get(`${BASE_URL}/api/v1/products/${productId}`, { headers: DEFAULT_HEADERS });
    checkResponse(res, 'product detail');
    productDetailDuration.add(Date.now() - t0);

    const isCached = res.headers['X-Cache'] === 'HIT' || res.headers['x-cache'] === 'HIT';
    cacheHitRate.add(isCached);
    sleep_random(1500, 4000);
  });

  // ── Related Products ────────────────────────────────────────────────────────
  group('Related Products', () => {
    const productId = randomItem(PRODUCT_IDS);
    const res = http.get(
      `${BASE_URL}/api/v1/products/${productId}/related?limit=8`,
      { headers: DEFAULT_HEADERS }
    );
    checkResponse(res, 'related products');
    sleep_random(500, 1500);
  });

  // ── Pagination ──────────────────────────────────────────────────────────────
  group('Pagination', () => {
    const page = Math.floor(Math.random() * 5) + 2;
    const res = http.get(
      `${BASE_URL}/api/v1/products?page=${page}&size=24`,
      { headers: DEFAULT_HEADERS }
    );
    checkResponse(res, `page ${page}`);
    sleep_random(800, 2000);
  });

  sleep_random(500, 1500);
}
