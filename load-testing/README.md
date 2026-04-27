# Load Testing

Three complementary load testing frameworks targeting the ShopOS API gateway.

## Tools

| Tool | Use case | Entry point |
|---|---|---|
| k6 | Threshold-based SLO gating in CI | `k6/scripts/` |
| Locust | Exploratory load with live web UI | `locust/locustfile.py` |
| Gatling | JVM-based high concurrency + HTML reports | `gatling/pom.xml` |

---

## k6

### Prerequisites
```bash
# macOS
brew install k6

# Linux
sudo snap install k6

# Docker
docker pull grafana/k6
```

### Scripts

| Script | Description | VUs | Duration |
|---|---|---|---|
| `smoke.js` | Critical path sanity check | 1 | 2m |
| `product-browse.js` | Catalog browsing | 0â†’100 | ~10m |
| `search-load.js` | Search service load | 0â†’80 | ~10m |
| `checkout-flow.js` | Full purchase journey | 0â†’50 | ~10m |
| `spike-test.js` | Sudden burst (auto-scale validation) | 0â†’500 | ~5m |
| `soak.js` | Sustained load (memory leak detection) | 30 | 2h |

### Run

```bash
# Smoke
k6 run k6/scripts/smoke.js

# Load test with custom base URL
BASE_URL=https://api.staging.shopos.io k6 run k6/scripts/checkout-flow.js

# Soak (30m in CI, 2h overnight)
K6_SOAK_DURATION=30m K6_SOAK_VUS=20 k6 run k6/scripts/soak.js

# Output to InfluxDB (Grafana dashboard)
k6 run --out influxdb=http://influxdb:8086/k6 k6/scripts/checkout-flow.js
```

---

## Locust

### Prerequisites
```bash
pip install locust
```

### Run

```bash
cd locust

# Interactive web UI (http://localhost:8089)
locust -f locustfile.py --host http://localhost:8080

# Headless with profile
locust -f locustfile.py \
  --headless \
  --users 50 \
  --spawn-rate 5 \
  --run-time 15m \
  --host http://api-gateway.platform.svc.cluster.local:8080

# Distributed (1 master + N workers)
locust -f locustfile.py --master
locust -f locustfile.py --worker --master-host=<master-ip>
```

### User classes

| Class | Behaviour | Target mix |
|---|---|---|
| `BrowseUser` | Anonymous read (FastHttp) | 60% |
| `ShopperUser` | Authenticated browse + cart | 30% |
| `PowerUser` | Full purchase flow | 10% |

---

## Gatling

### Prerequisites
- Java 11+
- Maven 3.8+

### Run

```bash
cd gatling

# Commerce simulation (default)
mvn gatling:test

# Search simulation
mvn gatling:test -Dgatling.simulationClass=com.shopos.SearchSimulation

# Custom target
mvn gatling:test -DBASE_URL=https://api.staging.shopos.io

# HTML report at: target/gatling-results/
```

### Simulations

| Class | Description |
|---|---|
| `CommerceSimulation` | Full purchase journey: login â†’ browse â†’ cart â†’ checkout â†’ payment |
| `SearchSimulation` | Search service: keyword, filtered, autocomplete, facets |

---

## CI Integration

k6 runs as a blocking CI step. Failing thresholds fail the pipeline.

```yaml
# Example Drone step
- name: load-test-smoke
  image: grafana/k6
  environment:
    BASE_URL:
      from_secret: staging_api_url
  commands:
    - k6 run load-testing/k6/scripts/smoke.js
```

---

## Thresholds Reference

See [`k6/config/thresholds.json`](k6/config/thresholds.json) for all baseline SLO definitions.

| Metric | Smoke | Load | Spike | Soak |
|---|---|---|---|---|
| `http_req_failed` | <1% | <5% | <15% | <1% |
| `p95 latency` | <2s | <3s | <10s | <3s |
| `p99 latency` | â€” | <5s | <15s | <5s |
| `payment_success_rate` | â€” | >95% | â€” | â€” |
