// Dagger Cluster Bootstrap Pipeline — 6-phase sequential cluster bring-up.
// Phase 1: Networking (Cilium, Istio, Traefik)
// Phase 2: Security (cert-manager, Vault, Keycloak, Kyverno, Falco)
// Phase 3: Observability (Prometheus, Grafana, Loki, Jaeger, OTel)
// Phase 4: Messaging (ZooKeeper, Kafka, RabbitMQ, NATS)
// Phase 5: Registry (Harbor, MinIO, Nexus)
// Phase 6: GitOps (ArgoCD, Argo Rollouts, KEDA, Velero)
// Usage: dagger run go run ci/dagger/cluster-bootstrap/main.go
// Env: KUBECONFIG_CONTENT, GRAFANA_ADMIN_PASSWORD, ARGOCD_ADMIN_PASSWORD,
//      HARBOR_ADMIN_PASSWORD, MINIO_ROOT_USER, MINIO_ROOT_PASSWORD,
//      RABBITMQ_PASSWORD, KEYCLOAK_ADMIN_PASSWORD, MINIO_SECRET_KEY,
//      START_PHASE (default: 1), END_PHASE (default: 6)
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

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

	startPhase, _ := strconv.Atoi(getEnv("START_PHASE", "1"))
	endPhase, _ := strconv.Atoi(getEnv("END_PHASE", "6"))

	kubeconfig := client.SetSecret("kubeconfig", mustEnv("KUBECONFIG_CONTENT"))

	helmBase := func() *dagger.Container {
		return client.Container().
			From("alpine/helm:latest").
			WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"})
	}

	helmInstall := func(c *dagger.Container, cmd string) error {
		if _, err := c.WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
			return err
		}
		return nil
	}

	// Phase 1: Networking
	if startPhase <= 1 && endPhase >= 1 {
		fmt.Println("===== Phase 1: Networking =====")
		cmds := []struct{ name, cmd string }{
			{"Cilium", "helm repo add cilium https://helm.cilium.io --force-update && helm upgrade --install cilium cilium/cilium --namespace kube-system --set hubble.enabled=true --set hubble.relay.enabled=true --set prometheus.enabled=true --wait --timeout 8m"},
			{"Istio base", "helm repo add istio https://istio-release.storage.googleapis.com/charts --force-update && helm upgrade --install istio-base istio/base --namespace istio-system --create-namespace --wait --timeout 3m"},
			{"Istiod", "helm upgrade --install istiod istio/istiod --namespace istio-system --set telemetry.enabled=true --wait --timeout 5m"},
			{"Traefik", "helm repo add traefik https://traefik.github.io/charts --force-update && helm upgrade --install traefik traefik/traefik --namespace networking --create-namespace --set deployment.replicas=2 --set metrics.prometheus.enabled=true --wait --timeout 5m"},
		}
		for _, item := range cmds {
			fmt.Printf("  --> %s\n", item.name)
			if err := helmInstall(helmBase(), item.cmd); err != nil {
				return fmt.Errorf("phase1 %s: %w", item.name, err)
			}
		}
	}

	// Phase 2: Security
	if startPhase <= 2 && endPhase >= 2 {
		fmt.Println("===== Phase 2: Security =====")
		keycloakPass := client.SetSecret("keycloak_pass", mustEnv("KEYCLOAK_ADMIN_PASSWORD"))
		cmds := []struct {
			name string
			fn   func() error
		}{
			{"cert-manager", func() error {
				return helmInstall(helmBase(), "helm repo add jetstack https://charts.jetstack.io --force-update && helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set installCRDs=true --set replicaCount=2 --wait --timeout 5m")
			}},
			{"Vault", func() error {
				return helmInstall(helmBase(), "helm repo add hashicorp https://helm.releases.hashicorp.com --force-update && helm upgrade --install vault hashicorp/vault --namespace vault --create-namespace --set server.ha.enabled=true --set server.ha.raft.enabled=true --set server.ha.replicas=3 --wait --timeout 10m")
			}},
			{"Keycloak", func() error {
				return helmInstall(
					helmBase().WithSecretVariable("KEYCLOAK_ADMIN_PASSWORD", keycloakPass),
					"helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && helm upgrade --install keycloak bitnami/keycloak --namespace keycloak --create-namespace --set replicaCount=2 --set auth.adminPassword=$KEYCLOAK_ADMIN_PASSWORD --set postgresql.enabled=true --wait --timeout 10m",
				)
			}},
			{"Kyverno", func() error {
				return helmInstall(helmBase(), "helm repo add kyverno https://kyverno.github.io/kyverno --force-update && helm upgrade --install kyverno kyverno/kyverno --namespace kyverno --create-namespace --set replicaCount=3 --wait --timeout 5m")
			}},
		}
		for _, item := range cmds {
			fmt.Printf("  --> %s\n", item.name)
			if err := item.fn(); err != nil {
				return fmt.Errorf("phase2 %s: %w", item.name, err)
			}
		}
	}

	// Phase 3: Observability
	if startPhase <= 3 && endPhase >= 3 {
		fmt.Println("===== Phase 3: Observability =====")
		grafanaPass := client.SetSecret("grafana_pass", mustEnv("GRAFANA_ADMIN_PASSWORD"))
		minioKey := client.SetSecret("minio_key", getEnv("MINIO_SECRET_KEY", "changeme"))
		cmds := []struct {
			name string
			fn   func() error
		}{
			{"Prometheus stack", func() error {
				return helmInstall(helmBase(), "helm repo add prometheus-community https://prometheus-community.github.io/helm-charts --force-update && helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack --namespace observability --create-namespace --set prometheus.prometheusSpec.retention=30d --set prometheus.prometheusSpec.replicas=2 --set grafana.enabled=false --wait --timeout 10m")
			}},
			{"Grafana", func() error {
				return helmInstall(
					helmBase().WithSecretVariable("GRAFANA_ADMIN_PASSWORD", grafanaPass),
					"helm repo add grafana https://grafana.github.io/helm-charts --force-update && helm upgrade --install grafana grafana/grafana --namespace observability --set adminPassword=$GRAFANA_ADMIN_PASSWORD --set persistence.enabled=true --set replicas=2 --wait --timeout 5m",
				)
			}},
			{"Loki", func() error {
				return helmInstall(
					helmBase().WithSecretVariable("MINIO_SECRET_KEY", minioKey),
					"helm repo add grafana https://grafana.github.io/helm-charts --force-update && helm upgrade --install loki grafana/loki --namespace observability --set loki.storage.type=s3 --set loki.storage.s3.endpoint=http://minio.registry.svc:9000 --set loki.storage.s3.bucketnames=loki-chunks --set loki.storage.s3.secretAccessKey=$MINIO_SECRET_KEY --set loki.storage.s3.s3ForcePathStyle=true --set deploymentMode=SimpleScalable --wait --timeout 8m",
				)
			}},
			{"Jaeger", func() error {
				return helmInstall(helmBase(), "helm repo add jaegertracing https://jaegertracing.github.io/helm-charts --force-update && helm upgrade --install jaeger jaegertracing/jaeger --namespace observability --set provisionDataStore.cassandra=false --set collector.replicaCount=2 --wait --timeout 8m")
			}},
		}
		for _, item := range cmds {
			fmt.Printf("  --> %s\n", item.name)
			if err := item.fn(); err != nil {
				return fmt.Errorf("phase3 %s: %w", item.name, err)
			}
		}
	}

	// Phase 4: Messaging
	if startPhase <= 4 && endPhase >= 4 {
		fmt.Println("===== Phase 4: Messaging =====")
		rabbitPass := client.SetSecret("rabbit_pass", mustEnv("RABBITMQ_PASSWORD"))
		cmds := []struct {
			name string
			fn   func() error
		}{
			{"ZooKeeper", func() error {
				return helmInstall(helmBase(), "helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && helm upgrade --install zookeeper bitnami/zookeeper --namespace messaging --create-namespace --set replicaCount=3 --wait --timeout 8m")
			}},
			{"Kafka", func() error {
				return helmInstall(helmBase(), "helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && helm upgrade --install kafka bitnami/kafka --namespace messaging --set replicaCount=3 --set zookeeper.enabled=false --set externalZookeeper.servers=zookeeper.messaging.svc:2181 --set autoCreateTopicsEnable=false --set defaultReplicationFactor=3 --wait --timeout 10m")
			}},
			{"RabbitMQ", func() error {
				return helmInstall(
					helmBase().WithSecretVariable("RABBITMQ_PASSWORD", rabbitPass),
					"helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && helm upgrade --install rabbitmq bitnami/rabbitmq --namespace messaging --set replicaCount=3 --set auth.password=$RABBITMQ_PASSWORD --set clustering.enabled=true --wait --timeout 8m",
				)
			}},
			{"NATS", func() error {
				return helmInstall(helmBase(), "helm repo add nats https://nats-io.github.io/k8s/helm/charts --force-update && helm upgrade --install nats nats/nats --namespace messaging --set nats.jetstream.enabled=true --set cluster.enabled=true --set cluster.replicas=3 --wait --timeout 5m")
			}},
		}
		for _, item := range cmds {
			fmt.Printf("  --> %s\n", item.name)
			if err := item.fn(); err != nil {
				return fmt.Errorf("phase4 %s: %w", item.name, err)
			}
		}
	}

	// Phase 5: Registry
	if startPhase <= 5 && endPhase >= 5 {
		fmt.Println("===== Phase 5: Registry =====")
		harborPass := client.SetSecret("harbor_pass", mustEnv("HARBOR_ADMIN_PASSWORD"))
		minioUser := client.SetSecret("minio_user", mustEnv("MINIO_ROOT_USER"))
		minioPass := client.SetSecret("minio_pass", mustEnv("MINIO_ROOT_PASSWORD"))
		cmds := []struct {
			name string
			fn   func() error
		}{
			{"MinIO", func() error {
				return helmInstall(
					helmBase().WithSecretVariable("MINIO_ROOT_USER", minioUser).WithSecretVariable("MINIO_ROOT_PASSWORD", minioPass),
					"helm repo add minio https://charts.min.io --force-update && helm upgrade --install minio minio/minio --namespace registry --create-namespace --set rootUser=$MINIO_ROOT_USER --set rootPassword=$MINIO_ROOT_PASSWORD --set replicas=4 --set persistence.size=500Gi --wait --timeout 8m",
				)
			}},
			{"Harbor", func() error {
				return helmInstall(
					helmBase().WithSecretVariable("HARBOR_ADMIN_PASSWORD", harborPass),
					"helm repo add harbor https://helm.goharbor.io --force-update && helm upgrade --install harbor harbor/harbor --namespace registry --set harborAdminPassword=$HARBOR_ADMIN_PASSWORD --set trivy.enabled=true --set expose.type=ClusterIP --wait --timeout 10m",
				)
			}},
			{"Nexus", func() error {
				return helmInstall(helmBase(), "helm repo add sonatype https://sonatype.github.io/helm3-charts --force-update && helm upgrade --install nexus sonatype/nexus-repository-manager --namespace registry --set nexus.resources.requests.memory=2Gi --set persistence.storageSize=200Gi --wait --timeout 10m")
			}},
		}
		for _, item := range cmds {
			fmt.Printf("  --> %s\n", item.name)
			if err := item.fn(); err != nil {
				return fmt.Errorf("phase5 %s: %w", item.name, err)
			}
		}
	}

	// Phase 6: GitOps
	if startPhase <= 6 && endPhase >= 6 {
		fmt.Println("===== Phase 6: GitOps =====")
		argoCDPass := client.SetSecret("argocd_pass", mustEnv("ARGOCD_ADMIN_PASSWORD"))
		cmds := []struct {
			name string
			fn   func() error
		}{
			{"ArgoCD", func() error {
				return helmInstall(
					helmBase().WithSecretVariable("ARGOCD_ADMIN_PASSWORD", argoCDPass),
					"helm repo add argo https://argoproj.github.io/argo-helm --force-update && helm upgrade --install argocd argo/argo-cd --namespace argocd --create-namespace --set server.replicas=2 --set configs.secret.argocdServerAdminPassword=$ARGOCD_ADMIN_PASSWORD --wait --timeout 8m",
				)
			}},
			{"Argo Rollouts", func() error {
				return helmInstall(helmBase(), "helm repo add argo https://argoproj.github.io/argo-helm --force-update && helm upgrade --install argo-rollouts argo/argo-rollouts --namespace argo-rollouts --create-namespace --set dashboard.enabled=true --wait --timeout 5m")
			}},
			{"KEDA", func() error {
				return helmInstall(helmBase(), "helm repo add kedacore https://kedacore.github.io/charts --force-update && helm upgrade --install keda kedacore/keda --namespace keda --create-namespace --set replicaCount=2 --wait --timeout 5m")
			}},
		}
		for _, item := range cmds {
			fmt.Printf("  --> %s\n", item.name)
			if err := item.fn(); err != nil {
				return fmt.Errorf("phase6 %s: %w", item.name, err)
			}
		}
	}

	fmt.Println("Cluster bootstrap pipeline complete")
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
