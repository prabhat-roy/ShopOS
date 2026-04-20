// Dagger Observability Stack Pipeline — Prometheus, Grafana, Loki, Jaeger, OTel.
// Usage: dagger run go run ci/dagger/observability/main.go
// Env: KUBECONFIG_CONTENT, ACTION (INSTALL|UNINSTALL),
//      GRAFANA_ADMIN_PASSWORD, MINIO_SECRET_KEY
package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	action := getEnv("ACTION", "INSTALL")
	kubeconfig := client.SetSecret("kubeconfig", mustEnv("KUBECONFIG_CONTENT"))
	src := client.Host().Directory(".")

	helmBase := func() *dagger.Container {
		return client.Container().
			From("alpine/helm:latest").
			WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"})
	}

	// kube-prometheus-stack
	fmt.Println("==> kube-prometheus-stack")
	cmd := "helm repo add prometheus-community https://prometheus-community.github.io/helm-charts --force-update && " +
		"helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack " +
		"--namespace observability --create-namespace " +
		"--set prometheus.prometheusSpec.retention=30d " +
		"--set prometheus.prometheusSpec.replicas=2 " +
		"--set grafana.enabled=false --wait --timeout 10m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall kube-prometheus-stack -n observability 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("prometheus stack: %w", err)
	}

	// Grafana
	fmt.Println("==> Grafana")
	grafanaPass := client.SetSecret("grafana_pass", mustEnv("GRAFANA_ADMIN_PASSWORD"))
	cmd = "helm repo add grafana https://grafana.github.io/helm-charts --force-update && " +
		"helm upgrade --install grafana grafana/grafana --namespace observability " +
		"--set adminPassword=$GRAFANA_ADMIN_PASSWORD --set persistence.enabled=true --set replicas=2 --wait --timeout 5m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall grafana -n observability 2>/dev/null || true"
	}
	c := helmBase().WithSecretVariable("GRAFANA_ADMIN_PASSWORD", grafanaPass).
		WithExec([]string{"sh", "-c", cmd})
	if action != "UNINSTALL" {
		c = c.WithMountedDirectory("/workspace", src).
			WithExec([]string{"sh", "-c", "kubectl apply -f /workspace/observability/grafana/ 2>/dev/null || true"})
	}
	if _, err := c.Stdout(ctx); err != nil {
		return fmt.Errorf("grafana: %w", err)
	}

	// Loki
	fmt.Println("==> Loki")
	minioKey := client.SetSecret("minio_key", getEnv("MINIO_SECRET_KEY", "changeme"))
	cmd = "helm repo add grafana https://grafana.github.io/helm-charts --force-update && " +
		"helm upgrade --install loki grafana/loki --namespace observability " +
		"--set loki.storage.type=s3 " +
		"--set loki.storage.s3.endpoint=http://minio.registry.svc:9000 " +
		"--set loki.storage.s3.bucketnames=loki-chunks " +
		"--set loki.storage.s3.secretAccessKey=$MINIO_SECRET_KEY " +
		"--set loki.storage.s3.s3ForcePathStyle=true " +
		"--set deploymentMode=SimpleScalable --wait --timeout 8m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall loki -n observability 2>/dev/null || true"
	}
	if _, err := helmBase().WithSecretVariable("MINIO_SECRET_KEY", minioKey).
		WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("loki: %w", err)
	}

	// Jaeger
	fmt.Println("==> Jaeger")
	cmd = "helm repo add jaegertracing https://jaegertracing.github.io/helm-charts --force-update && " +
		"helm upgrade --install jaeger jaegertracing/jaeger --namespace observability " +
		"--set provisionDataStore.cassandra=false " +
		"--set collector.replicaCount=2 --set query.replicaCount=2 --wait --timeout 8m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall jaeger -n observability 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("jaeger: %w", err)
	}

	// OTel Collector
	fmt.Println("==> OTel Collector")
	cmd = "helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts --force-update && " +
		"helm upgrade --install otel-collector open-telemetry/opentelemetry-collector --namespace observability " +
		"--set mode=daemonset " +
		"--set config.exporters.otlp.endpoint=jaeger-collector.observability.svc:4317 --wait --timeout 5m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall otel-collector -n observability 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("otel-collector: %w", err)
	}

	fmt.Println("Observability pipeline complete")
	return nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "required env var %s is not set\n", key)
		os.Exit(1)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
