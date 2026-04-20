// Dagger Networking Infrastructure Pipeline — Cilium, Istio, Traefik, Consul.
// Usage: dagger run go run ci/dagger/networking/main.go
// Env: KUBECONFIG_CONTENT, ACTION (INSTALL|UNINSTALL),
//      CILIUM, ISTIO, TRAEFIK, CONSUL
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

	helmBase := func() *dagger.Container {
		return client.Container().
			From("alpine/helm:latest").
			WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"})
	}

	// Cilium
	if getEnv("CILIUM", "true") == "true" {
		fmt.Println("==> Cilium CNI")
		cmd := "helm repo add cilium https://helm.cilium.io --force-update && " +
			"helm upgrade --install cilium cilium/cilium --namespace kube-system " +
			"--set hubble.enabled=true --set hubble.relay.enabled=true " +
			"--set hubble.ui.enabled=true --set prometheus.enabled=true --wait --timeout 8m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall cilium -n kube-system 2>/dev/null || true"
		}
		if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return fmt.Errorf("cilium: %w", err)
		}
	}

	// Istio
	if getEnv("ISTIO", "true") == "true" {
		fmt.Println("==> Istio service mesh")
		if action == "UNINSTALL" {
			for _, rel := range []string{"istio-ingress -n istio-ingress", "istiod -n istio-system", "istio-base -n istio-system"} {
				helmBase().WithExec([]string{"sh", "-c", "helm uninstall " + rel + " 2>/dev/null || true"}).Stdout(ctx)
			}
		} else {
			cmds := []string{
				"helm repo add istio https://istio-release.storage.googleapis.com/charts --force-update",
				"helm upgrade --install istio-base istio/base --namespace istio-system --create-namespace --wait --timeout 3m",
				"helm upgrade --install istiod istio/istiod --namespace istio-system --set telemetry.enabled=true --set pilot.traceSampling=1.0 --wait --timeout 5m",
				"helm upgrade --install istio-ingress istio/gateway --namespace istio-ingress --create-namespace --wait --timeout 5m",
			}
			c := helmBase()
			for _, cmd := range cmds {
				c = c.WithExec([]string{"sh", "-c", cmd})
			}
			if _, err := c.Stdout(ctx); err != nil {
				return fmt.Errorf("istio: %w", err)
			}
		}
	}

	// Traefik
	if getEnv("TRAEFIK", "true") == "true" {
		fmt.Println("==> Traefik edge router")
		cmd := "helm repo add traefik https://traefik.github.io/charts --force-update && " +
			"helm upgrade --install traefik traefik/traefik --namespace networking --create-namespace " +
			"--set deployment.replicas=2 --set ingressClass.enabled=true --set ingressClass.isDefaultClass=true " +
			"--set metrics.prometheus.enabled=true --wait --timeout 5m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall traefik -n networking 2>/dev/null || true"
		}
		if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return fmt.Errorf("traefik: %w", err)
		}
	}

	// Consul
	if getEnv("CONSUL", "false") == "true" {
		fmt.Println("==> Consul service discovery")
		cmd := "helm repo add hashicorp https://helm.releases.hashicorp.com --force-update && " +
			"helm upgrade --install consul hashicorp/consul --namespace consul --create-namespace " +
			"--set global.datacenter=shopos-dc1 --set server.replicas=3 --set connectInject.enabled=true --wait --timeout 8m"
		if action == "UNINSTALL" {
			cmd = "helm uninstall consul -n consul 2>/dev/null || true"
		}
		if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return fmt.Errorf("consul: %w", err)
		}
	}

	fmt.Println("Networking pipeline complete")
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
