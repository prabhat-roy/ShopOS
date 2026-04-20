// Dagger GitOps Pipeline — ArgoCD, Argo Rollouts, KEDA, Velero, ArgoCD bootstrap.
// Usage: dagger run go run ci/dagger/gitops/main.go
// Env: KUBECONFIG_CONTENT, ACTION (INSTALL|UNINSTALL),
//      ARGOCD_ADMIN_PASSWORD, GITOPS_REPO_URL, MINIO_ACCESS_KEY, MINIO_SECRET_KEY
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

	kubectlBase := func() *dagger.Container {
		return client.Container().
			From("bitnami/kubectl:latest").
			WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"})
	}

	// ArgoCD
	fmt.Println("==> ArgoCD")
	argoCDPass := client.SetSecret("argocd_pass", mustEnv("ARGOCD_ADMIN_PASSWORD"))
	cmd := "helm repo add argo https://argoproj.github.io/argo-helm --force-update && " +
		"helm upgrade --install argocd argo/argo-cd --namespace argocd --create-namespace " +
		"--set server.replicas=2 --set repoServer.replicas=2 " +
		"--set configs.secret.argocdServerAdminPassword=$ARGOCD_ADMIN_PASSWORD --wait --timeout 8m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall argocd -n argocd 2>/dev/null || true"
	}
	if _, err := helmBase().
		WithSecretVariable("ARGOCD_ADMIN_PASSWORD", argoCDPass).
		WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("argocd: %w", err)
	}

	// Argo Rollouts
	fmt.Println("==> Argo Rollouts")
	cmd = "helm repo add argo https://argoproj.github.io/argo-helm --force-update && " +
		"helm upgrade --install argo-rollouts argo/argo-rollouts --namespace argo-rollouts --create-namespace " +
		"--set dashboard.enabled=true --wait --timeout 5m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall argo-rollouts -n argo-rollouts 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("argo-rollouts: %w", err)
	}

	// Argo Workflows
	fmt.Println("==> Argo Workflows")
	cmd = "helm repo add argo https://argoproj.github.io/argo-helm --force-update && " +
		"helm upgrade --install argo-workflows argo/argo-workflows --namespace argo --create-namespace " +
		"--set server.replicas=2 --wait --timeout 5m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall argo-workflows -n argo 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("argo-workflows: %w", err)
	}

	// KEDA
	fmt.Println("==> KEDA")
	cmd = "helm repo add kedacore https://kedacore.github.io/charts --force-update && " +
		"helm upgrade --install keda kedacore/keda --namespace keda --create-namespace " +
		"--set replicaCount=2 --wait --timeout 5m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall keda -n keda 2>/dev/null || true"
	}
	c := helmBase().WithExec([]string{"sh", "-c", cmd})
	if action != "UNINSTALL" {
		c = c.WithMountedDirectory("/workspace", src).
			WithExec([]string{"sh", "-c", "kubectl apply -f /workspace/kubernetes/keda/ 2>/dev/null || true"})
	}
	if _, err := c.Stdout(ctx); err != nil {
		return fmt.Errorf("keda: %w", err)
	}

	// Velero
	fmt.Println("==> Velero")
	minioKey := client.SetSecret("minio_access", mustEnv("MINIO_ACCESS_KEY"))
	minioSecret := client.SetSecret("minio_secret", mustEnv("MINIO_SECRET_KEY"))
	cmd = "helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts --force-update && " +
		"helm upgrade --install velero vmware-tanzu/velero --namespace velero --create-namespace " +
		"--set configuration.provider=aws " +
		"--set configuration.backupStorageLocation.bucket=velero-backups " +
		"--set configuration.backupStorageLocation.config.region=minio " +
		"--set configuration.backupStorageLocation.config.s3ForcePathStyle=true " +
		"--set configuration.backupStorageLocation.config.s3Url=http://minio.registry.svc:9000 " +
		"--set credentials.secretContents.cloud=\"[default]\\naws_access_key_id=$MINIO_ACCESS_KEY\\naws_secret_access_key=$MINIO_SECRET_KEY\" " +
		"--set initContainers[0].name=velero-plugin-for-aws " +
		"--set initContainers[0].image=velero/velero-plugin-for-aws:v1.10.0 " +
		"--set initContainers[0].volumeMounts[0].mountPath=/target " +
		"--set initContainers[0].volumeMounts[0].name=plugins --wait --timeout 5m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall velero -n velero 2>/dev/null || true"
	}
	if _, err := helmBase().
		WithSecretVariable("MINIO_ACCESS_KEY", minioKey).
		WithSecretVariable("MINIO_SECRET_KEY", minioSecret).
		WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("velero: %w", err)
	}

	// ArgoCD App-of-Apps bootstrap
	if action != "UNINSTALL" {
		repoURL := getEnv("GITOPS_REPO_URL", "")
		if repoURL != "" {
			fmt.Println("==> ArgoCD App-of-Apps bootstrap")
			if _, err := kubectlBase().
				WithMountedDirectory("/workspace", src).
				WithExec([]string{"sh", "-c",
					"kubectl apply -f /workspace/gitops/argocd/ -n argocd 2>/dev/null || true"}).
				Stdout(ctx); err != nil {
				fmt.Println("ArgoCD bootstrap: non-blocking error:", err)
			}
		}
	}

	fmt.Println("GitOps pipeline complete")
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
