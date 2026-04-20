def call() {
    def svc     = env.TEST_SERVICE
    def url     = env.SERVICE_URL
    def profile = env.LOAD_PROFILE ?: 'medium'
    def domain  = env.TEST_DOMAIN

    sh 'mkdir -p reports/load/k6'

    def profiles = [
        light : [vus: 10,  duration: '2m',  rampUp: '30s', rampDown: '30s'],
        medium: [vus: 50,  duration: '5m',  rampUp: '1m',  rampDown: '30s'],
        heavy : [vus: 200, duration: '10m', rampUp: '2m',  rampDown: '1m'],
        spike : [vus: 500, duration: '1m',  rampUp: '30s', rampDown: '30s'],
    ]
    def p   = profiles[profile] ?: profiles.medium
    def vus = env.LOAD_VUS?.trim() ? env.LOAD_VUS.toInteger() : p.vus
    def dur = env.LOAD_DURATION?.trim() ?: p.duration

    echo "k6 profile=${profile} vus=${vus} duration=${dur}"

    // Write k6 scripts inline
    sh """
        mkdir -p /tmp/k6-scripts

        cat > /tmp/k6-scripts/checkout-flow.js << 'JSEOF'
import http from 'k6/http';
import { sleep, check } from 'k6';
import { SharedArray } from 'k6/data';

const products = new SharedArray('products', function () {
  return [
    { id: 'prod-001', sku: 'LAPTOP-X1',  price: 999.99 },
    { id: 'prod-002', sku: 'PHONE-S21',  price: 699.99 },
    { id: 'prod-003', sku: 'TABLET-P10', price: 499.99 },
    { id: 'prod-004', sku: 'WATCH-W3',   price: 299.99 },
    { id: 'prod-005', sku: 'EARBUDS-E2', price: 149.99 },
  ];
});

export const options = {
  stages: [
    { duration: '1m',  target: 20 },
    { duration: '5m',  target: 20 },
    { duration: '1m',  target: 0  },
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://api-gateway.shopos-platform.svc.cluster.local:8080';
const HEADERS  = { 'Content-Type': 'application/json' };

export default function () {
  const product = products[Math.floor(Math.random() * products.length)];

  const cartRes = http.post(
    BASE_URL + '/api/v1/cart',
    JSON.stringify({ productId: product.id, quantity: 1 }),
    { headers: HEADERS }
  );
  check(cartRes, { 'cart: status 200 or 201': (r) => r.status === 200 || r.status === 201 });
  sleep(1);

  const checkoutRes = http.post(
    BASE_URL + '/api/v1/checkout',
    JSON.stringify({
      cartId:          cartRes.json('cartId') || 'cart-test-id',
      paymentMethod:   'CREDIT_CARD',
      shippingAddress: { line1: '123 Main St', city: 'Anytown', country: 'US', zip: '10001' },
    }),
    { headers: HEADERS }
  );
  check(checkoutRes, { 'checkout: status 200 or 201': (r) => r.status === 200 || r.status === 201 });
  sleep(1);

  const orderId = checkoutRes.json('orderId') || 'order-test-id';
  const orderRes = http.get(BASE_URL + '/api/v1/orders/' + orderId);
  check(orderRes, { 'order: status 200': (r) => r.status === 200 });
  sleep(2);
}
JSEOF

        cat > /tmp/k6-scripts/product-browse.js << 'JSEOF'
import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '2m',  target: 50 },
    { duration: '30s', target: 0  },
  ],
};

const BASE_URL = __ENV.BASE_URL || 'http://api-gateway.shopos-platform.svc.cluster.local:8080';

export default function () {
  const res = http.get(BASE_URL + '/api/v1/products?page=1&limit=20');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
JSEOF

        cat > /tmp/k6-scripts/search-load.js << 'JSEOF'
import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '3m',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    'http_req_duration{scenario:constant_request_rate}': ['p(99)<500'],
    'http_req_failed{scenario:constant_request_rate}': ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://api-gateway.shopos-platform.svc.cluster.local:8080';

export default function () {
  const res = http.get(BASE_URL + '/api/v1/search?q=laptop&limit=10');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(0.1);
}
JSEOF
    """

    def scripts = []
    if (domain == 'commerce') {
        scripts += ['/tmp/k6-scripts/checkout-flow.js']
    }
    if (domain == 'catalog') {
        scripts += ['/tmp/k6-scripts/product-browse.js', '/tmp/k6-scripts/search-load.js']
    }
    if (scripts.isEmpty()) {
        scripts += ['/tmp/k6-scripts/product-browse.js']
    }

    scripts.each { script ->
        def scriptName = script.tokenize('/')[-1].replace('.js', '')
        sh """
            echo "=== k6: ${scriptName} (${vus} VUs × ${dur}) ==="
            docker run --rm \
                --network host \
                -v /tmp/k6-scripts:/scripts \
                -v \${WORKSPACE}/reports/load/k6:/reports \
                -e BASE_URL=${url} \
                grafana/k6:latest run \
                --vus ${vus} \
                --duration ${dur} \
                --out json=/reports/${scriptName}.json \
                --summary-export /reports/${scriptName}-summary.json \
                /scripts/${scriptName}.js || true
        """
    }

    if (profile != 'spike') {
        sh """
            echo "=== k6: Spike test (500 VUs × 30s) ==="
            docker run --rm \
                --network host \
                -v /tmp/k6-scripts:/scripts \
                -v \${WORKSPACE}/reports/load/k6:/reports \
                -e BASE_URL=${url} \
                grafana/k6:latest run \
                --stage 0s:0,30s:500,1m:500,90s:0 \
                --out json=/reports/spike.json \
                --summary-export /reports/spike-summary.json \
                /scripts/checkout-flow.js || true
        """
    }

    sh 'rm -rf /tmp/k6-scripts'
    echo 'k6 load tests complete — reports/load/k6/'
}
return this
