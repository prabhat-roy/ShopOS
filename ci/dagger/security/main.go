// Dagger Security Infrastructure Pipeline — installs security tools onto a K8s cluster.
// Usage: dagger run go run ci/dagger/security/main.go
// Env: KUBECONFIG_CONTENT, ACTION (INSTALL|UNINSTALL),
//      CERT_MANAGER, VAULT, KEYCLOAK, KEYCLOAK_ADMIN_PASSWORD,
//      KYVERNO, FALCO, OPA, SLACK_WEBHOOK
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

	// cert-manager
	if getEnv("CERT_MANAGER", "true") == "true" {
		fmt.Println("==> cert-manager")
		cmd := "helm repo add jetstack https://charts.jetstack.io --force-update && " +
			"helm upgrade --install cert-manager jetstack/cert-manager " +
			"--namespace cert-manager --create-namespace " +
			"--set installCRDs=true --set replicaCount=2 --wait --timeout 5m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall cert-manager -n cert-manager 2>/dev/null || true"
		}
		if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return fmt.Errorf("cert-manager: %w", err)
		}
	}

	// Vault
	if getEnv("VAULT", "true") == "true" {
		fmt.Println("==> Vault HA")
		cmd := "helm repo add hashicorp https://helm.releases.hashicorp.com --force-update && " +
			"helm upgrade --install vault hashicorp/vault " +
			"--namespace vault --create-namespace " +
			"--set server.ha.enabled=true --set server.ha.raft.enabled=true " +
			"--set server.ha.replicas=3 --wait --timeout 10m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall vault -n vault 2>/dev/null || true"
		}
		if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return fmt.Errorf("vault: %w", err)
		}
	}

	// Keycloak
	if getEnv("KEYCLOAK", "true") == "true" {
		fmt.Println("==> Keycloak")
		keycloakPass := client.SetSecret("keycloak_pass", mustEnv("KEYCLOAK_ADMIN_PASSWORD"))
		cmd := "helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && " +
			"helm upgrade --install keycloak bitnami/keycloak " +
			"--namespace keycloak --create-namespace " +
			"--set replicaCount=2 --set auth.adminPassword=$KEYCLOAK_ADMIN_PASSWORD " +
			"--set postgresql.enabled=true --wait --timeout 10m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall keycloak -n keycloak 2>/dev/null || true"
		}
		if _, err := helmBase().
			WithSecretVariable("KEYCLOAK_ADMIN_PASSWORD", keycloakPass).
			WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return fmt.Errorf("keycloak: %w", err)
		}
	}

	// Kyverno
	if getEnv("KYVERNO", "true") == "true" {
		fmt.Println("==> Kyverno")
		cmd := "helm repo add kyverno https://kyverno.github.io/kyverno --force-update && " +
			"helm upgrade --install kyverno kyverno/kyverno " +
			"--namespace kyverno --create-namespace " +
			"--set replicaCount=3 --wait --timeout 5m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall kyverno -n kyverno 2>/dev/null || true"
		}
		c := helmBase().WithMountedDirectory("/workspace", src).WithExec([]string{"sh", "-c", cmd})
		if action != "UNINSTALL" {
			c = c.WithExec([]string{"sh", "-c", "kubectl apply -f /workspace/security/kyverno/ 2>/dev/null || true"})
		}
		if _, err := c.Stdout(ctx); err != nil {
			return fmt.Errorf("kyverno: %w", err)
		}
	}

	// Falco
	if getEnv("FALCO", "true") == "true" {
		fmt.Println("==> Falco")
		slackWebhook := client.SetSecret("slack_webhook", getEnv("SLACK_WEBHOOK", ""))
		cmd := "helm repo add falcosecurity https://falcosecurity.github.io/charts --force-update && " +
			"helm upgrade --install falco falcosecurity/falco " +
			"--namespace falco --create-namespace " +
			"--set driver.kind=ebpf " +
			"--set falcosidekick.enabled=true " +
			"--set falcosidekick.config.slack.webhookurl=$SLACK_WEBHOOK " +
			"--wait --timeout 8m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall falco -n falco 2>/dev/null || true"
		}
		if _, err := helmBase().
			WithSecretVariable("SLACK_WEBHOOK", slackWebhook).
			WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return fmt.Errorf("falco: %w", err)
		}
	}

	fmt.Println("Security pipeline complete")
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
