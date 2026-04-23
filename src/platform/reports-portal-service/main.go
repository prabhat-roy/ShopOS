package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	// Observability
	GrafanaURL         string
	PrometheusURL      string
	AlertmanagerURL    string
	JaegerURL          string
	ZipkinURL          string
	TempoURL           string
	LokiURL            string
	KibanaURL          string
	OpenSearchDashURL  string
	PyroscopeURL       string
	SigNozURL          string
	GlitchTipURL       string
	UptimeKumaURL      string

	// CI/CD & Build
	JenkinsURL         string
	ArgoCDURL          string
	ArgoWorkflowsURL   string
	FluxURL            string
	HarborURL          string
	NexusURL           string
	GiteaURL           string
	SonarQubeURL       string

	// Security
	DefectDojoURL      string
	DependencyTrackURL string
	VaultURL           string
	KeycloakURL        string
	WazuhURL           string
	RekorURL           string

	// Contract & API Testing
	PactBrokerURL      string

	// Messaging
	AKHQUrl            string
	KafkaUIURL         string
	RabbitMQURL        string
	NATSMonURL         string
	ConduktorURL       string

	// Data
	AirflowURL         string
	PrefectURL         string
	DagsterURL         string
	MarquezURL         string
	SupersetURL        string
	MLflowURL          string

	// Platform
	BackstageURL       string
	TemporalURL        string
	UnleashURL         string
	ChaosURL           string
	LitmusURL          string

	// Reports Portal itself
	JenkinsBuildURL    string
	Port               string
}

func loadConfig() Config {
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	return Config{
		GrafanaURL:         get("GRAFANA_URL", "http://grafana-grafana.grafana.svc.cluster.local:3000"),
		PrometheusURL:      get("PROMETHEUS_URL", "http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local:9090"),
		AlertmanagerURL:    get("ALERTMANAGER_URL", "http://prometheus-kube-prometheus-alertmanager.monitoring.svc.cluster.local:9093"),
		JaegerURL:          get("JAEGER_URL", "http://jaeger-query.observability.svc.cluster.local:16686"),
		ZipkinURL:          get("ZIPKIN_URL", "http://zipkin.observability.svc.cluster.local:9411"),
		TempoURL:           get("TEMPO_URL", "http://tempo.observability.svc.cluster.local:3200"),
		LokiURL:            get("LOKI_URL", "http://loki-gateway.observability.svc.cluster.local:80"),
		KibanaURL:          get("KIBANA_URL", "http://kibana.observability.svc.cluster.local:5601"),
		OpenSearchDashURL:  get("OPENSEARCH_DASHBOARDS_URL", "http://opensearch-dashboards.observability.svc.cluster.local:5601"),
		PyroscopeURL:       get("PYROSCOPE_URL", "http://pyroscope.observability.svc.cluster.local:4040"),
		SigNozURL:          get("SIGNOZ_URL", "http://signoz-frontend.observability.svc.cluster.local:3301"),
		GlitchTipURL:       get("GLITCHTIP_URL", "http://glitchtip.observability.svc.cluster.local:8000"),
		UptimeKumaURL:      get("UPTIME_KUMA_URL", "http://uptime-kuma.observability.svc.cluster.local:3001"),
		JenkinsURL:         get("JENKINS_URL", "http://jenkins.ci.svc.cluster.local:8080"),
		ArgoCDURL:          get("ARGOCD_URL", "http://argocd-server.argocd.svc.cluster.local:80"),
		ArgoWorkflowsURL:   get("ARGO_WORKFLOWS_URL", "http://argo-workflows-server.argo.svc.cluster.local:2746"),
		FluxURL:            get("FLUX_URL", "http://weave-gitops.flux-system.svc.cluster.local:9001"),
		HarborURL:          get("HARBOR_URL", "http://harbor.registry.svc.cluster.local:80"),
		NexusURL:           get("NEXUS_URL", "http://nexus-nexus-repository-manager.registry.svc.cluster.local:8081"),
		GiteaURL:           get("GITEA_URL", "http://gitea.registry.svc.cluster.local:3000"),
		SonarQubeURL:       get("SONARQUBE_URL", "http://sonarqube.security.svc.cluster.local:9000"),
		DefectDojoURL:      get("DEFECTDOJO_URL", "http://defectdojo.security.svc.cluster.local:8080"),
		DependencyTrackURL: get("DEPENDENCY_TRACK_URL", "http://dependency-track.security.svc.cluster.local:8080"),
		VaultURL:           get("VAULT_URL", "http://vault.security.svc.cluster.local:8200"),
		KeycloakURL:        get("KEYCLOAK_URL", "http://keycloak.security.svc.cluster.local:8080"),
		WazuhURL:           get("WAZUH_URL", "http://wazuh-dashboard.security.svc.cluster.local:5601"),
		RekorURL:           get("REKOR_URL", "http://rekor-server.security.svc.cluster.local:3000"),
		PactBrokerURL:      get("PACT_BROKER_URL", "http://pact-broker.platform.svc.cluster.local:9292"),
		AKHQUrl:            get("AKHQ_URL", "http://akhq.messaging.svc.cluster.local:8080"),
		KafkaUIURL:         get("KAFKA_UI_URL", "http://kafka-ui.messaging.svc.cluster.local:8080"),
		RabbitMQURL:        get("RABBITMQ_URL", "http://rabbitmq.messaging.svc.cluster.local:15672"),
		NATSMonURL:         get("NATS_MON_URL", "http://nats.messaging.svc.cluster.local:8222"),
		ConduktorURL:       get("CONDUKTOR_URL", "http://conduktor-platform.messaging.svc.cluster.local:8080"),
		AirflowURL:         get("AIRFLOW_URL", "http://airflow-webserver.data.svc.cluster.local:8080"),
		PrefectURL:         get("PREFECT_URL", "http://prefect-server.data.svc.cluster.local:4200"),
		DagsterURL:         get("DAGSTER_URL", "http://dagster-dagit.data.svc.cluster.local:3000"),
		MarquezURL:         get("MARQUEZ_URL", "http://marquez-web.data.svc.cluster.local:3000"),
		SupersetURL:        get("SUPERSET_URL", "http://superset.data.svc.cluster.local:8088"),
		MLflowURL:          get("MLFLOW_URL", "http://mlflow.ml.svc.cluster.local:5000"),
		BackstageURL:       get("BACKSTAGE_URL", "http://backstage.platform.svc.cluster.local:7007"),
		TemporalURL:        get("TEMPORAL_URL", "http://temporal-web.platform.svc.cluster.local:8080"),
		UnleashURL:         get("UNLEASH_URL", "http://unleash.platform.svc.cluster.local:4242"),
		ChaosURL:           get("CHAOS_MESH_URL", "http://chaos-dashboard.chaos-mesh.svc.cluster.local:2333"),
		LitmusURL:          get("LITMUS_URL", "http://litmusportal-frontend.litmus.svc.cluster.local:9091"),
		JenkinsBuildURL:    get("JENKINS_URL", "http://jenkins.ci.svc.cluster.local:8080"),
		Port:               get("PORT", "8300"),
	}
}

type ReportEntry struct {
	Name     string
	URL      string
	Badge    string
	Category string
}

var htmlTmpl = template.Must(template.New("portal").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>ShopOS — Reports Portal</title>
<style>
  :root {
    --bg: #0f1117; --surface: #1a1d27; --surface2: #22263a;
    --border: #2e3350; --text: #e2e8f0; --muted: #64748b;
    --accent: #6366f1; --green: #22c55e; --red: #ef4444;
    --yellow: #f59e0b; --blue: #3b82f6; --purple: #a855f7;
    --cyan: #06b6d4; --orange: #f97316;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: var(--bg); color: var(--text); font-family: 'Segoe UI', system-ui, sans-serif; min-height: 100vh; }
  header { background: var(--surface); border-bottom: 1px solid var(--border); padding: 1rem 2rem; display: flex; align-items: center; gap: 1rem; }
  header h1 { font-size: 1.5rem; font-weight: 700; }
  header .badge { background: var(--accent); color: #fff; padding: 0.2rem 0.6rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; }
  header .ts { margin-left: auto; color: var(--muted); font-size: 0.8rem; }
  nav { background: var(--surface2); border-bottom: 1px solid var(--border); padding: 0 2rem; display: flex; gap: 0; overflow-x: auto; }
  nav a { color: var(--muted); text-decoration: none; padding: 0.75rem 1rem; font-size: 0.85rem; white-space: nowrap; border-bottom: 2px solid transparent; transition: all 0.15s; }
  nav a:hover { color: var(--text); border-bottom-color: var(--accent); }
  main { padding: 2rem; max-width: 1600px; margin: 0 auto; }
  .section { margin-bottom: 2.5rem; }
  .section-title { font-size: 1rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-bottom: 1rem; display: flex; align-items: center; gap: 0.5rem; }
  .section-title::after { content: ''; flex: 1; height: 1px; background: var(--border); }
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
  .card { background: var(--surface); border: 1px solid var(--border); border-radius: 0.5rem; padding: 1rem 1.25rem; display: flex; flex-direction: column; gap: 0.5rem; transition: border-color 0.15s, transform 0.15s; }
  .card:hover { border-color: var(--accent); transform: translateY(-1px); }
  .card-header { display: flex; align-items: center; justify-content: space-between; }
  .card-name { font-weight: 600; font-size: 0.95rem; }
  .dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
  .dot-green { background: var(--green); box-shadow: 0 0 6px var(--green); }
  .dot-gray  { background: var(--muted); }
  .card-url { color: var(--muted); font-size: 0.75rem; word-break: break-all; }
  .card-link { margin-top: 0.25rem; }
  .card-link a { display: inline-flex; align-items: center; gap: 0.35rem; color: var(--accent); text-decoration: none; font-size: 0.82rem; font-weight: 500; }
  .card-link a:hover { text-decoration: underline; }
  .tag { display: inline-block; padding: 0.15rem 0.5rem; border-radius: 9999px; font-size: 0.7rem; font-weight: 600; }
  .tag-obs   { background: #1e3a5f; color: var(--blue); }
  .tag-sec   { background: #3b1515; color: var(--red); }
  .tag-ci    { background: #1a2e1a; color: var(--green); }
  .tag-msg   { background: #2d1f3e; color: var(--purple); }
  .tag-data  { background: #1f2d3a; color: var(--cyan); }
  .tag-plat  { background: #2a2010; color: var(--orange); }
  .tag-test  { background: #1e2a1e; color: #86efac; }
  .summary { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 0.75rem; margin-bottom: 2rem; }
  .stat { background: var(--surface); border: 1px solid var(--border); border-radius: 0.5rem; padding: 1rem; text-align: center; }
  .stat-num { font-size: 2rem; font-weight: 800; }
  .stat-label { color: var(--muted); font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; margin-top: 0.25rem; }
  .report-card { background: var(--surface); border: 1px solid var(--border); border-radius: 0.5rem; padding: 1.25rem; }
  .report-card h3 { font-size: 0.9rem; font-weight: 700; margin-bottom: 0.75rem; }
  .report-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; }
  .report-list li a { color: var(--accent); text-decoration: none; font-size: 0.85rem; }
  .report-list li a:hover { text-decoration: underline; }
  .report-list li { display: flex; align-items: baseline; gap: 0.4rem; }
  .report-list li::before { content: '▸'; color: var(--muted); font-size: 0.7rem; flex-shrink: 0; }
  footer { text-align: center; color: var(--muted); font-size: 0.75rem; padding: 2rem; border-top: 1px solid var(--border); margin-top: 2rem; }
</style>
</head>
<body>
<header>
  <span style="font-size:1.5rem">🛒</span>
  <h1>ShopOS Reports Portal</h1>
  <span class="badge">Enterprise Platform</span>
  <span class="ts">Last refreshed: {{.Timestamp}}</span>
</header>
<nav>
  <a href="#summary">Summary</a>
  <a href="#build-reports">Build Reports</a>
  <a href="#security-reports">Security</a>
  <a href="#quality-reports">Code Quality</a>
  <a href="#test-reports">Tests</a>
  <a href="#performance-reports">Performance</a>
  <a href="#observability">Observability</a>
  <a href="#gitops">GitOps / CI-CD</a>
  <a href="#messaging">Messaging</a>
  <a href="#data">Data & ML</a>
  <a href="#platform">Platform</a>
  <a href="#all-dashboards">All Dashboards</a>
</nav>
<main>

<div id="summary" class="section">
  <div class="section-title">Platform Summary</div>
  <div class="summary">
    <div class="stat"><div class="stat-num" style="color:var(--accent)">262</div><div class="stat-label">Total Services</div></div>
    <div class="stat"><div class="stat-num" style="color:var(--green)">22</div><div class="stat-label">Domains</div></div>
    <div class="stat"><div class="stat-num" style="color:var(--blue)">19</div><div class="stat-label">Languages</div></div>
    <div class="stat"><div class="stat-num" style="color:var(--purple)">15</div><div class="stat-label">CI Platforms</div></div>
    <div class="stat"><div class="stat-num" style="color:var(--cyan)">50+</div><div class="stat-label">Security Tools</div></div>
    <div class="stat"><div class="stat-num" style="color:var(--orange)">35+</div><div class="stat-label">Observability Tools</div></div>
  </div>
</div>

<div id="build-reports" class="section">
  <div class="section-title">Build & CI Reports</div>
  <div class="grid">
    <div class="report-card">
      <h3>Jenkins Build Reports</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/01-install-tools/lastBuild/console" target="_blank">Install Tools — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/02-cluster-bootstrap/lastBuild/console" target="_blank">Cluster Bootstrap — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/console" target="_blank">Pre-Deploy (CI) — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/04-deploy/lastBuild/console" target="_blank">Deploy (GitOps) — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/console" target="_blank">Post-Deploy (Validation) — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/06-security/lastBuild/console" target="_blank">Security Install — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/07-observability/lastBuild/console" target="_blank">Observability Install — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/console" target="_blank">API Quality — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/09-tooling/lastBuild/console" target="_blank">Tooling — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}/job/10-gitops/lastBuild/console" target="_blank">GitOps — Last Build Log</a></li>
        <li><a href="{{.Jenkins}}" target="_blank">All Jenkins Pipelines →</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Build Artifacts (Jenkins)</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/" target="_blank">Pre-Deploy Artifacts (SBOM, SARIF)</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/" target="_blank">Post-Deploy Artifacts (ZAP, Nuclei)</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/artifact/" target="_blank">API Quality Artifacts (Spectral, Terrascan)</a></li>
        <li><a href="{{.Jenkins}}/job/06-security/lastBuild/artifact/" target="_blank">Security Tool Artifacts</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Test Results (JUnit in Jenkins)</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/testReport/" target="_blank">Post-Deploy Test Report</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/testReport/" target="_blank">API Quality Test Report (Hurl + Spectral)</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/testReport/" target="_blank">Pre-Deploy Unit Tests</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>ArgoCD / GitOps Status</h3>
      <ul class="report-list">
        <li><a href="{{.ArgoCD}}/applications" target="_blank">All Applications (230)</a></li>
        <li><a href="{{.ArgoCD}}/applications?health=Degraded" target="_blank">Degraded Applications</a></li>
        <li><a href="{{.ArgoCD}}/applications?sync=OutOfSync" target="_blank">Out-of-Sync Applications</a></li>
        <li><a href="{{.ArgoWorkflows}}" target="_blank">Argo Workflows Runs</a></li>
      </ul>
    </div>
  </div>
</div>

<div id="security-reports" class="section">
  <div class="section-title">Security Reports</div>
  <div class="grid">
    <div class="report-card">
      <h3>DefectDojo — All Findings</h3>
      <ul class="report-list">
        <li><a href="{{.DefectDojo}}/finding" target="_blank">All Findings</a></li>
        <li><a href="{{.DefectDojo}}/finding?severity=Critical" target="_blank">Critical Findings</a></li>
        <li><a href="{{.DefectDojo}}/finding?severity=High" target="_blank">High Severity</a></li>
        <li><a href="{{.DefectDojo}}/engagement" target="_blank">Engagement List</a></li>
        <li><a href="{{.DefectDojo}}/reports/auto" target="_blank">Auto-Generated Reports</a></li>
        <li><a href="{{.DefectDojo}}" target="_blank">DefectDojo Dashboard →</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Dependency-Track — SBOM & CVEs</h3>
      <ul class="report-list">
        <li><a href="{{.DependencyTrack}}" target="_blank">Projects Overview</a></li>
        <li><a href="{{.DependencyTrack}}/#/vulnerabilities" target="_blank">All Vulnerabilities</a></li>
        <li><a href="{{.DependencyTrack}}/#/projects" target="_blank">SBOM Projects</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Jenkins — Security Scan Artifacts</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/trivy-image-report.json" target="_blank">Trivy Image Scan (JSON)</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/grype-report.json" target="_blank">Grype CVE Report (JSON)</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/sbom.json" target="_blank">SBOM (CycloneDX JSON)</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/semgrep-results.sarif" target="_blank">Semgrep SAST (SARIF)</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/gitleaks-report.json" target="_blank">GitLeaks Secrets Scan</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/trufflehog-report.json" target="_blank">TruffleHog Scan</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/checkov-results.sarif" target="_blank">Checkov IaC (SARIF)</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/kics-results.sarif" target="_blank">KICS IaC (SARIF)</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/artifact/terrascan-aws.sarif" target="_blank">Terrascan AWS (SARIF)</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>DAST Reports</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/zap-report.html" target="_blank">OWASP ZAP HTML Report</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/zap-report.json" target="_blank">OWASP ZAP JSON</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/nuclei-results.json" target="_blank">Nuclei CVE Scan (JSON)</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/nuclei-results.sarif" target="_blank">Nuclei CVE Scan (SARIF)</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Wazuh SIEM</h3>
      <ul class="report-list">
        <li><a href="{{.Wazuh}}/app/wazuh#/overview" target="_blank">Security Overview</a></li>
        <li><a href="{{.Wazuh}}/app/wazuh#/vulnerability-detector" target="_blank">Vulnerability Detector</a></li>
        <li><a href="{{.Wazuh}}/app/wazuh#/compliance" target="_blank">Compliance (PCI, GDPR)</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Supply Chain Integrity</h3>
      <ul class="report-list">
        <li><a href="{{.Rekor}}" target="_blank">Rekor Transparency Log</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/cosign-verify.txt" target="_blank">Cosign Verify Output</a></li>
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/notation-verify.txt" target="_blank">Notation Verify Output</a></li>
      </ul>
    </div>
  </div>
</div>

<div id="quality-reports" class="section">
  <div class="section-title">Code Quality Reports</div>
  <div class="grid">
    <div class="report-card">
      <h3>SonarQube — Static Analysis</h3>
      <ul class="report-list">
        <li><a href="{{.SonarQube}}/projects" target="_blank">All Projects</a></li>
        <li><a href="{{.SonarQube}}/issues" target="_blank">All Issues</a></li>
        <li><a href="{{.SonarQube}}/security_hotspots" target="_blank">Security Hotspots</a></li>
        <li><a href="{{.SonarQube}}/dashboard?id=shopos" target="_blank">ShopOS Quality Gate</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>OpenAPI / Spectral Lint</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/testReport/" target="_blank">Spectral JUnit Results</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/artifact/spectral-results.xml" target="_blank">spectral-results.xml</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>License Compliance</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/03-pre-deploy/lastBuild/artifact/license-report.json" target="_blank">License Report (JSON)</a></li>
        <li><a href="{{.DependencyTrack}}/#/projects" target="_blank">Component Licenses (Dependency-Track)</a></li>
      </ul>
    </div>
  </div>
</div>

<div id="test-reports" class="section">
  <div class="section-title">Test Reports</div>
  <div class="grid">
    <div class="report-card">
      <h3>Hurl — HTTP API Tests</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/testReport/" target="_blank">Post-Deploy Hurl Results</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/testReport/" target="_blank">API Quality Hurl Results</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/artifact/api-testing/hurl/hurl-health-results.xml" target="_blank">Health Check Results (XML)</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/artifact/api-testing/hurl/hurl-auth-results.xml" target="_blank">Auth Flow Results (XML)</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/artifact/api-testing/hurl/hurl-catalog-results.xml" target="_blank">Catalog Flow Results (XML)</a></li>
        <li><a href="{{.Jenkins}}/job/08-api-quality/lastBuild/artifact/api-testing/hurl/hurl-checkout-results.xml" target="_blank">Checkout Flow Results (XML)</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Pact — Contract Tests</h3>
      <ul class="report-list">
        <li><a href="{{.PactBroker}}" target="_blank">Pact Broker Dashboard</a></li>
        <li><a href="{{.PactBroker}}/pacts" target="_blank">All Pacts</a></li>
        <li><a href="{{.PactBroker}}/matrix" target="_blank">Compatibility Matrix</a></li>
        <li><a href="{{.PactBroker}}/verifications" target="_blank">Verifications</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>E2E / Integration Tests</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/testReport/" target="_blank">Integration Test Results</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/smoke-test-results.xml" target="_blank">Smoke Test Results</a></li>
      </ul>
    </div>
  </div>
</div>

<div id="performance-reports" class="section">
  <div class="section-title">Performance & Load Test Reports</div>
  <div class="grid">
    <div class="report-card">
      <h3>Gatling Reports</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/gatling/" target="_blank">Gatling HTML Report (Jenkins Publisher)</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/load-testing/gatling/results/" target="_blank">Gatling Raw Results</a></li>
        <li><a href="{{.Grafana}}/d/gatling/gatling-metrics?orgId=1" target="_blank">Gatling Metrics Dashboard (Grafana)</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Locust Reports</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/locust-report.html" target="_blank">Locust HTML Report</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/locust-stats.csv" target="_blank">Locust Stats (CSV)</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>k6 Reports</h3>
      <ul class="report-list">
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/k6-results.json" target="_blank">k6 Results (JSON)</a></li>
        <li><a href="{{.Grafana}}/d/k6/k6-load-testing-results?orgId=1" target="_blank">k6 Grafana Dashboard</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Chaos Engineering</h3>
      <ul class="report-list">
        <li><a href="{{.Chaos}}" target="_blank">Chaos Mesh Dashboard</a></li>
        <li><a href="{{.Litmus}}" target="_blank">Litmus Center</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/chaos-results.txt" target="_blank">Chaos Run Results</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>SLO Reports</h3>
      <ul class="report-list">
        <li><a href="{{.Grafana}}/d/slo-overview/slo-overview?orgId=1" target="_blank">SLO Overview (Grafana)</a></li>
        <li><a href="{{.Grafana}}/d/pyrra/pyrra-slo-dashboard?orgId=1" target="_blank">Pyrra SLO Dashboard</a></li>
        <li><a href="{{.Jenkins}}/job/05-post-deploy/lastBuild/artifact/slo-validation.txt" target="_blank">SLO Validation Output</a></li>
      </ul>
    </div>
    <div class="report-card">
      <h3>Profiling</h3>
      <ul class="report-list">
        <li><a href="{{.Pyroscope}}" target="_blank">Pyroscope — Continuous Profiling</a></li>
        <li><a href="{{.Pyroscope}}/?query=process_cpu%7Bservice_name%3D%22order-service%22%7D" target="_blank">Order Service CPU Profile</a></li>
      </ul>
    </div>
  </div>
</div>

<div id="observability" class="section">
  <div class="section-title">Observability Dashboards</div>
  <div class="grid">
    <div class="card"><div class="card-header"><span class="card-name">Grafana</span><span class="tag tag-obs">metrics</span></div><div class="card-url">{{.Grafana}}</div><div class="card-link"><a href="{{.Grafana}}/dashboards" target="_blank">↗ Open Dashboards</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Prometheus</span><span class="tag tag-obs">metrics</span></div><div class="card-url">{{.Prometheus}}</div><div class="card-link"><a href="{{.Prometheus}}/targets" target="_blank">↗ Scrape Targets</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Alertmanager</span><span class="tag tag-obs">alerts</span></div><div class="card-url">{{.Alertmanager}}</div><div class="card-link"><a href="{{.Alertmanager}}/#/alerts" target="_blank">↗ Active Alerts</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Jaeger</span><span class="tag tag-obs">tracing</span></div><div class="card-url">{{.Jaeger}}</div><div class="card-link"><a href="{{.Jaeger}}/search" target="_blank">↗ Search Traces</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Zipkin</span><span class="tag tag-obs">tracing</span></div><div class="card-url">{{.Zipkin}}</div><div class="card-link"><a href="{{.Zipkin}}" target="_blank">↗ Open Zipkin</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Kibana</span><span class="tag tag-obs">logs</span></div><div class="card-url">{{.Kibana}}</div><div class="card-link"><a href="{{.Kibana}}/app/discover" target="_blank">↗ Discover Logs</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">OpenSearch Dashboards</span><span class="tag tag-obs">logs</span></div><div class="card-url">{{.OpenSearchDash}}</div><div class="card-link"><a href="{{.OpenSearchDash}}" target="_blank">↗ Open</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Pyroscope</span><span class="tag tag-obs">profiling</span></div><div class="card-url">{{.Pyroscope}}</div><div class="card-link"><a href="{{.Pyroscope}}" target="_blank">↗ Flame Graphs</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">SigNoz</span><span class="tag tag-obs">full-stack</span></div><div class="card-url">{{.SigNoz}}</div><div class="card-link"><a href="{{.SigNoz}}" target="_blank">↗ Open SigNoz</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">GlitchTip</span><span class="tag tag-obs">errors</span></div><div class="card-url">{{.GlitchTip}}</div><div class="card-link"><a href="{{.GlitchTip}}" target="_blank">↗ Error Tracker</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Uptime Kuma</span><span class="tag tag-obs">uptime</span></div><div class="card-url">{{.UptimeKuma}}</div><div class="card-link"><a href="{{.UptimeKuma}}" target="_blank">↗ Status Page</a></div></div>
  </div>
</div>

<div id="gitops" class="section">
  <div class="section-title">GitOps & CI-CD</div>
  <div class="grid">
    <div class="card"><div class="card-header"><span class="card-name">Jenkins</span><span class="tag tag-ci">ci-cd</span></div><div class="card-url">{{.Jenkins}}</div><div class="card-link"><a href="{{.Jenkins}}" target="_blank">↗ All Pipelines</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">ArgoCD</span><span class="tag tag-ci">gitops</span></div><div class="card-url">{{.ArgoCD}}</div><div class="card-link"><a href="{{.ArgoCD}}/applications" target="_blank">↗ Applications (230)</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Argo Workflows</span><span class="tag tag-ci">workflows</span></div><div class="card-url">{{.ArgoWorkflows}}</div><div class="card-link"><a href="{{.ArgoWorkflows}}" target="_blank">↗ Workflow Runs</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Flux / Weave GitOps</span><span class="tag tag-ci">gitops</span></div><div class="card-url">{{.Flux}}</div><div class="card-link"><a href="{{.Flux}}" target="_blank">↗ Open Weave GitOps</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Harbor Registry</span><span class="tag tag-ci">registry</span></div><div class="card-url">{{.Harbor}}</div><div class="card-link"><a href="{{.Harbor}}" target="_blank">↗ Container Images</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Nexus</span><span class="tag tag-ci">registry</span></div><div class="card-url">{{.Nexus}}</div><div class="card-link"><a href="{{.Nexus}}" target="_blank">↗ Artifact Repository</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Gitea</span><span class="tag tag-ci">git</span></div><div class="card-url">{{.Gitea}}</div><div class="card-link"><a href="{{.Gitea}}" target="_blank">↗ Self-hosted Git</a></div></div>
  </div>
</div>

<div id="messaging" class="section">
  <div class="section-title">Messaging & Streaming</div>
  <div class="grid">
    <div class="card"><div class="card-header"><span class="card-name">AKHQ (Kafka UI)</span><span class="tag tag-msg">kafka</span></div><div class="card-url">{{.AKHQ}}</div><div class="card-link"><a href="{{.AKHQ}}" target="_blank">↗ Topic Browser</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Kafka UI</span><span class="tag tag-msg">kafka</span></div><div class="card-url">{{.KafkaUI}}</div><div class="card-link"><a href="{{.KafkaUI}}" target="_blank">↗ Open Kafka UI</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">RabbitMQ Management</span><span class="tag tag-msg">rabbitmq</span></div><div class="card-url">{{.RabbitMQ}}</div><div class="card-link"><a href="{{.RabbitMQ}}" target="_blank">↗ Queue Manager</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">NATS Monitoring</span><span class="tag tag-msg">nats</span></div><div class="card-url">{{.NATSMon}}</div><div class="card-link"><a href="{{.NATSMon}}" target="_blank">↗ NATS Stats</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Conduktor</span><span class="tag tag-msg">kafka</span></div><div class="card-url">{{.Conduktor}}</div><div class="card-link"><a href="{{.Conduktor}}" target="_blank">↗ Kafka Governance</a></div></div>
  </div>
</div>

<div id="data" class="section">
  <div class="section-title">Data, Analytics & ML</div>
  <div class="grid">
    <div class="card"><div class="card-header"><span class="card-name">Apache Airflow</span><span class="tag tag-data">orchestration</span></div><div class="card-url">{{.Airflow}}</div><div class="card-link"><a href="{{.Airflow}}/dags" target="_blank">↗ DAGs</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Prefect</span><span class="tag tag-data">orchestration</span></div><div class="card-url">{{.Prefect}}</div><div class="card-link"><a href="{{.Prefect}}" target="_blank">↗ Flow Runs</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Dagster</span><span class="tag tag-data">orchestration</span></div><div class="card-url">{{.Dagster}}</div><div class="card-link"><a href="{{.Dagster}}" target="_blank">↗ Asset Graph</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Marquez (Data Lineage)</span><span class="tag tag-data">lineage</span></div><div class="card-url">{{.Marquez}}</div><div class="card-link"><a href="{{.Marquez}}" target="_blank">↗ Lineage Graph</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Apache Superset</span><span class="tag tag-data">bi</span></div><div class="card-url">{{.Superset}}</div><div class="card-link"><a href="{{.Superset}}/dashboard/list" target="_blank">↗ BI Dashboards</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">MLflow</span><span class="tag tag-data">ml</span></div><div class="card-url">{{.MLflow}}</div><div class="card-link"><a href="{{.MLflow}}/#/experiments" target="_blank">↗ Experiments</a></div></div>
  </div>
</div>

<div id="platform" class="section">
  <div class="section-title">Platform Engineering</div>
  <div class="grid">
    <div class="card"><div class="card-header"><span class="card-name">Backstage</span><span class="tag tag-plat">portal</span></div><div class="card-url">{{.Backstage}}</div><div class="card-link"><a href="{{.Backstage}}/catalog" target="_blank">↗ Service Catalog</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Temporal UI</span><span class="tag tag-plat">workflows</span></div><div class="card-url">{{.Temporal}}</div><div class="card-link"><a href="{{.Temporal}}" target="_blank">↗ Workflow Executions</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Unleash</span><span class="tag tag-plat">feature-flags</span></div><div class="card-url">{{.Unleash}}</div><div class="card-link"><a href="{{.Unleash}}" target="_blank">↗ Feature Toggles</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Vault</span><span class="tag tag-sec">secrets</span></div><div class="card-url">{{.Vault}}</div><div class="card-link"><a href="{{.Vault}}/ui" target="_blank">↗ Secrets Manager</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Keycloak</span><span class="tag tag-sec">iam</span></div><div class="card-url">{{.Keycloak}}</div><div class="card-link"><a href="{{.Keycloak}}/admin" target="_blank">↗ Identity Admin</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Pact Broker</span><span class="tag tag-test">contracts</span></div><div class="card-url">{{.PactBroker}}</div><div class="card-link"><a href="{{.PactBroker}}" target="_blank">↗ Contract Registry</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Chaos Mesh</span><span class="tag tag-plat">chaos</span></div><div class="card-url">{{.Chaos}}</div><div class="card-link"><a href="{{.Chaos}}" target="_blank">↗ Chaos Experiments</a></div></div>
    <div class="card"><div class="card-header"><span class="card-name">Litmus</span><span class="tag tag-plat">chaos</span></div><div class="card-url">{{.Litmus}}</div><div class="card-link"><a href="{{.Litmus}}" target="_blank">↗ Chaos Scenarios</a></div></div>
  </div>
</div>

</main>
<footer>ShopOS Reports Portal — 262 services · 22 domains · 19 languages · 15 CI platforms · 50+ security tools · 35+ observability tools</footer>
</body>
</html>`))

type TemplateData struct {
	Timestamp      string
	Jenkins        string
	ArgoCD         string
	ArgoWorkflows  string
	Flux           string
	Harbor         string
	Nexus          string
	Gitea          string
	SonarQube      string
	DefectDojo     string
	DependencyTrack string
	Vault          string
	Keycloak       string
	Wazuh          string
	Rekor          string
	PactBroker     string
	Grafana        string
	Prometheus     string
	Alertmanager   string
	Jaeger         string
	Zipkin         string
	Tempo          string
	Loki           string
	Kibana         string
	OpenSearchDash string
	Pyroscope      string
	SigNoz         string
	GlitchTip      string
	UptimeKuma     string
	AKHQ           string
	KafkaUI        string
	RabbitMQ       string
	NATSMon        string
	Conduktor      string
	Airflow        string
	Prefect        string
	Dagster        string
	Marquez        string
	Superset       string
	MLflow         string
	Backstage      string
	Temporal       string
	Unleash        string
	Chaos          string
	Litmus         string
}

func main() {
	cfg := loadConfig()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		urls := map[string]string{
			"grafana":          cfg.GrafanaURL,
			"prometheus":       cfg.PrometheusURL,
			"jenkins":          cfg.JenkinsURL,
			"argocd":           cfg.ArgoCDURL,
			"sonarqube":        cfg.SonarQubeURL,
			"defectdojo":       cfg.DefectDojoURL,
			"pact_broker":      cfg.PactBrokerURL,
			"harbor":           cfg.HarborURL,
			"reports_portal":   fmt.Sprintf("http://reports-portal.platform.svc.cluster.local:%s", cfg.Port),
		}
		json.NewEncoder(w).Encode(urls)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Timestamp:      time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
			Jenkins:        cfg.JenkinsURL,
			ArgoCD:         cfg.ArgoCDURL,
			ArgoWorkflows:  cfg.ArgoWorkflowsURL,
			Flux:           cfg.FluxURL,
			Harbor:         cfg.HarborURL,
			Nexus:          cfg.NexusURL,
			Gitea:          cfg.GiteaURL,
			SonarQube:      cfg.SonarQubeURL,
			DefectDojo:     cfg.DefectDojoURL,
			DependencyTrack: cfg.DependencyTrackURL,
			Vault:          cfg.VaultURL,
			Keycloak:       cfg.KeycloakURL,
			Wazuh:          cfg.WazuhURL,
			Rekor:          cfg.RekorURL,
			PactBroker:     cfg.PactBrokerURL,
			Grafana:        cfg.GrafanaURL,
			Prometheus:     cfg.PrometheusURL,
			Alertmanager:   cfg.AlertmanagerURL,
			Jaeger:         cfg.JaegerURL,
			Zipkin:         cfg.ZipkinURL,
			Tempo:          cfg.TempoURL,
			Loki:           cfg.LokiURL,
			Kibana:         cfg.KibanaURL,
			OpenSearchDash: cfg.OpenSearchDashURL,
			Pyroscope:      cfg.PyroscopeURL,
			SigNoz:         cfg.SigNozURL,
			GlitchTip:      cfg.GlitchTipURL,
			UptimeKuma:     cfg.UptimeKumaURL,
			AKHQ:           cfg.AKHQUrl,
			KafkaUI:        cfg.KafkaUIURL,
			RabbitMQ:       cfg.RabbitMQURL,
			NATSMon:        cfg.NATSMonURL,
			Conduktor:      cfg.ConduktorURL,
			Airflow:        cfg.AirflowURL,
			Prefect:        cfg.PrefectURL,
			Dagster:        cfg.DagsterURL,
			Marquez:        cfg.MarquezURL,
			Superset:       cfg.SupersetURL,
			MLflow:         cfg.MLflowURL,
			Backstage:      cfg.BackstageURL,
			Temporal:       cfg.TemporalURL,
			Unleash:        cfg.UnleashURL,
			Chaos:          cfg.ChaosURL,
			Litmus:         cfg.LitmusURL,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := htmlTmpl.Execute(w, data); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})

	addr := ":" + cfg.Port
	log.Printf("reports-portal-service listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
