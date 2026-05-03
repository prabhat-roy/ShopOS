# Observability Stack — ShopOS

ShopOS implements the three pillars — metrics, traces, and logs — plus formal SLO tracking,
continuous profiling, error tracking, RUM, and uptime monitoring, all on a fully open-source
toolchain. All 303 services emit telemetry through the OpenTelemetry SDK; the OTel agent
DaemonSet collects and routes to the right backend per signal.

---

## Layout

```
observability/
├── otel-collector/          OTel agent (DaemonSet) + gateway (Deployment) with tail-sampling
├── prometheus/              Scrape config + per-domain alert rules
│   └── rules/               identity, catalog, financial, supply-chain, cx, comms, marketplace+b2b, infra, platform, commerce
├── alertmanager/            Routing tree (PagerDuty / Slack)
├── grafana/                 Provisioned dashboards + datasources
├── loki/                    Log aggregation (Loki)
├── tempo/                   Distributed tracing (Tempo)
├── jaeger/                  Distributed tracing (Jaeger — dev/staging)
├── zipkin/                  Lightweight tracing (legacy compat)
├── thanos/                  Long-term Prometheus storage (S3-backed)
├── victoria-metrics/        High-cardinality metrics
├── mimir/                   Multi-tenant Prometheus
├── victoria-logs/           High-volume log overflow (cheaper than Loki)
├── fluent-bit/              Log collection DaemonSet → Loki + VictoriaLogs split
├── fluentd/                 Heavier-weight log aggregator
├── opensearch/              Log search (security audit)
├── opensearch-dashboards/   Security dashboards
├── elasticsearch/           Log search (engineer debug)
├── kibana/                  Engineer log dashboards
├── logstash/                External log transformation
├── pyroscope/               Continuous CPU/memory profiling
├── parca/                   eBPF profiling (complement to Pyroscope)
├── pixie/                   eBPF auto-instrumentation (zero-code traces)
├── sentry/                  Error tracking (frontend)
├── glitchtip/               Error tracking (backend, Sentry-compatible)
├── openreplay/              Session replay for storefront
├── plausible/               Privacy-friendly web analytics
├── rum/                     OTel RUM config (Web Vitals + frontend errors)
├── uptime-kuma/             External uptime monitoring (feeds Cachet)
├── slo/                     Pyrra SLO definitions for tier-0 services
├── pyrra/                   Pyrra SLO controller
├── sloth/                   Sloth SLO rule generator
├── netdata/                 Per-second per-container metrics
├── kiali/                   Istio service mesh observability
├── perses/                  GitOps-native dashboards-as-code
├── signoz/                  Full-stack obs for analytics-ai team
├── komodor/                 K8s troubleshooting timelines per resource
├── healthchecks/            Self-hosted cron monitoring (alerts on miss)
├── quickwit/                S3-backed log search (cold logs)
└── opencost/                Per-namespace cost attribution
```

FinOps lives separately under [`finops/kubecost/`](../finops/kubecost/).

---

## Telemetry Pipeline

```
Services (OTel SDK + Prometheus exporter + JSON stdout logs)
   │
   ▼
OTel Collector agent (DaemonSet)         Fluent-bit (DaemonSet)
   │                                          │
   ├── traces  ──► Tempo (prod)               ├── high-volume logs ──► VictoriaLogs
   │              Jaeger (dev/staging)        └── tier-0 logs       ──► Loki
   │
   ├── metrics ──► Mimir (multi-tenant)
   │              Prometheus (short-term)
   │              Thanos (long-term S3)
   │              VictoriaMetrics (high-cardinality)
   │
   └── logs    ──► Loki (queryable in Grafana)

K8s events / kubelet stats ──► OTel hostmetrics + kubeletstats receivers
```

Tail-sampling decides which traces to keep:
- Always: errors, latency > 500ms, payment + financial domains
- Otherwise: 5% probabilistic sample

---

## Per-domain alerting

Each domain has its own Prometheus rules file in [`prometheus/rules/`](prometheus/rules/):
identity, catalog, financial, supply-chain, customer-experience, communications,
marketplace+b2b, infra, platform, commerce.

| Severity | Channel | Response |
|---|---|---|
| `critical` (label `page: pagerduty`) | PagerDuty primary on-call | Immediate |
| `warning` | Slack `#alerts-<team>` | 30 min |
| `info` | Slack `#ops` | Next business day |

---

## SLOs

Pyrra SLO definitions in [`slo/slo-definitions.yaml`](slo/slo-definitions.yaml) cover all
tier-0 services (api-gateway, checkout, payment, search, auth, order placement, storefront).
Pyrra emits multi-burn-rate Prometheus rules consumed by Alertmanager.

---

## Quick start

```bash
# Deploy stack via Helm umbrella + apply per-domain rules
helm upgrade --install shopos-observability helm/charts/observability -n monitoring --create-namespace

# Apply OTel agent + gateway
kubectl apply -f observability/otel-collector/

# Apply Fluent-bit DaemonSet
kubectl apply -f observability/fluent-bit/daemonset.yaml

# Apply per-domain Prometheus rules
kubectl apply -f observability/prometheus/rules/

# Apply Pyrra SLOs
kubectl apply -f observability/slo/

# Port-forward UIs
kubectl port-forward svc/grafana 3000:3000 -n monitoring         # http://localhost:3000
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
kubectl port-forward svc/jaeger-query 16686:16686 -n monitoring
kubectl port-forward svc/loki-gateway 3100:3100 -n monitoring
```

---

## Related

- Incident response (war room, Cachet status page): [`../incident/`](../incident/)
- FinOps (Kubecost): [`../finops/`](../finops/)
- Right-sizing recommendations: [`../kubernetes/scaling/vpa/`](../kubernetes/scaling/vpa/)
- Runbooks: [`../docs/runbooks/`](../docs/runbooks/)
