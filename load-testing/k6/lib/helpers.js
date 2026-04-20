import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

export const errorRate = new Rate('error_rate');
export const successRate = new Rate('success_rate');
export const requestDuration = new Trend('request_duration', true);
export const requestsFailed = new Counter('requests_failed');

export const BASE_URL = __ENV.BASE_URL || 'http://api-gateway.platform.svc.cluster.local:8080';
export const AUTH_URL = __ENV.AUTH_URL || 'http://auth-service.identity.svc.cluster.local:8080';

export const DEFAULT_HEADERS = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
  'X-Request-ID': () => `k6-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
};

export function checkResponse(res, name) {
  const success = check(res, {
    [`${name}: status 2xx`]: (r) => r.status >= 200 && r.status < 300,
    [`${name}: response time < 2s`]: (r) => r.timings.duration < 2000,
    [`${name}: has body`]: (r) => r.body && r.body.length > 0,
  });
  errorRate.add(!success);
  successRate.add(success);
  requestDuration.add(res.timings.duration);
  if (!success) {
    requestsFailed.add(1);
    console.error(`${name} failed: status=${res.status} duration=${res.timings.duration}ms`);
  }
  return success;
}

export function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function sleep_random(minMs, maxMs) {
  sleep((randomInt(minMs, maxMs)) / 1000);
}

export function getAuthToken(username = 'testuser@shopos.local', password = 'testpass') {
  const res = http.post(`${AUTH_URL}/api/v1/login`, JSON.stringify({ username, password }), {
    headers: DEFAULT_HEADERS,
  });
  if (res.status === 200) {
    return JSON.parse(res.body).access_token;
  }
  return null;
}

export function authHeaders(token) {
  return { ...DEFAULT_HEADERS, Authorization: `Bearer ${token}` };
}
