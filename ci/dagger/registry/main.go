// Dagger Registry Pipeline — Harbor, MinIO, Nexus, bucket provisioning.
// Usage: dagger run go run ci/dagger/registry/main.go
// Env: KUBECONFIG_CONTENT, ACTION (INSTALL|UNINSTALL),
//      HARBOR_ADMIN_PASSWORD, MINIO_ROOT_USER, MINIO_ROOT_PASSWORD
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	helmBase := func() *dagger.Container {
		return client.Container().
			From("alpine/helm:latest").
			WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"})
	}

	// Harbor
	fmt.Println("==> Harbor")
	harborPass := client.SetSecret("harbor_pass", mustEnv("HARBOR_ADMIN_PASSWORD"))
	cmd := "helm repo add harbor https://helm.goharbor.io --force-update && " +
		"helm upgrade --install harbor harbor/harbor --namespace registry --create-namespace " +
		"--set harborAdminPassword=$HARBOR_ADMIN_PASSWORD " +
		"--set persistence.persistentVolumeClaim.registry.size=100Gi " +
		"--set trivy.enabled=true --set notary.enabled=true " +
		"--set expose.type=ClusterIP --wait --timeout 10m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall harbor -n registry 2>/dev/null || true"
	}
	if _, err := helmBase().
		WithSecretVariable("HARBOR_ADMIN_PASSWORD", harborPass).
		WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("harbor: %w", err)
	}

	// MinIO
	fmt.Println("==> MinIO")
	minioUser := client.SetSecret("minio_user", mustEnv("MINIO_ROOT_USER"))
	minioPass := client.SetSecret("minio_pass", mustEnv("MINIO_ROOT_PASSWORD"))
	cmd = "helm repo add minio https://charts.min.io --force-update && " +
		"helm upgrade --install minio minio/minio --namespace registry " +
		"--set rootUser=$MINIO_ROOT_USER --set rootPassword=$MINIO_ROOT_PASSWORD " +
		"--set replicas=4 --set persistence.size=500Gi " +
		"--set resources.requests.memory=1Gi --wait --timeout 8m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall minio -n registry 2>/dev/null || true"
	}
	if _, err := helmBase().
		WithSecretVariable("MINIO_ROOT_USER", minioUser).
		WithSecretVariable("MINIO_ROOT_PASSWORD", minioPass).
		WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("minio: %w", err)
	}

	// Nexus
	fmt.Println("==> Nexus")
	cmd = "helm repo add sonatype https://sonatype.github.io/helm3-charts --force-update && " +
		"helm upgrade --install nexus sonatype/nexus-repository-manager --namespace registry " +
		"--set nexus.resources.requests.memory=2Gi --set nexus.resources.limits.memory=4Gi " +
		"--set persistence.storageSize=200Gi --wait --timeout 10m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall nexus -n registry 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("nexus: %w", err)
	}

	// Create MinIO buckets
	if action != "UNINSTALL" {
		fmt.Println("==> Creating MinIO buckets")
		buckets := []string{
			"harbor-registry", "velero-backups", "loki-chunks",
			"tempo-traces", "thanos", "mlflow", "artifacts", "media-assets",
		}
		cmds := []string{
			"mc alias set local http://minio.registry.svc:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD",
		}
		for _, b := range buckets {
			cmds = append(cmds, fmt.Sprintf("mc mb --ignore-existing local/%s", b))
		}
		script := strings.Join(cmds, " && ")
		if _, err := client.Container().
			From("minio/mc:latest").
			WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
			WithSecretVariable("MINIO_ROOT_USER", minioUser).
			WithSecretVariable("MINIO_ROOT_PASSWORD", minioPass).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"}).
			WithExec([]string{"sh", "-c", script}).
			Stdout(ctx); err != nil {
			fmt.Println("MinIO bucket creation: non-blocking error:", err)
		}
	}

	fmt.Println("Registry pipeline complete")
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
