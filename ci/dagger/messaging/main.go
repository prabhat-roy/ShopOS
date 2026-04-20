// Dagger Messaging Infrastructure Pipeline — ZooKeeper, Kafka, RabbitMQ, NATS.
// Usage: dagger run go run ci/dagger/messaging/main.go
// Env: KUBECONFIG_CONTENT, ACTION (INSTALL|UNINSTALL|CONFIGURE),
//      RABBITMQ_PASSWORD, CREATE_TOPICS
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

	// ZooKeeper
	fmt.Println("==> ZooKeeper")
	cmd := "helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && " +
		"helm upgrade --install zookeeper bitnami/zookeeper --namespace messaging --create-namespace " +
		"--set replicaCount=3 --wait --timeout 8m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall zookeeper -n messaging 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("zookeeper: %w", err)
	}

	// Kafka
	fmt.Println("==> Kafka")
	cmd = "helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && " +
		"helm upgrade --install kafka bitnami/kafka --namespace messaging " +
		"--set replicaCount=3 --set zookeeper.enabled=false " +
		"--set externalZookeeper.servers=zookeeper.messaging.svc:2181 " +
		"--set autoCreateTopicsEnable=false --set defaultReplicationFactor=3 " +
		"--set offsetsTopicReplicationFactor=3 --wait --timeout 10m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall kafka -n messaging 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("kafka: %w", err)
	}

	// RabbitMQ
	fmt.Println("==> RabbitMQ")
	rabbitPass := client.SetSecret("rabbitmq_pass", mustEnv("RABBITMQ_PASSWORD"))
	cmd = "helm repo add bitnami https://charts.bitnami.com/bitnami --force-update && " +
		"helm upgrade --install rabbitmq bitnami/rabbitmq --namespace messaging " +
		"--set replicaCount=3 --set auth.password=$RABBITMQ_PASSWORD --set clustering.enabled=true --wait --timeout 8m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall rabbitmq -n messaging 2>/dev/null || true"
	}
	if _, err := helmBase().
		WithSecretVariable("RABBITMQ_PASSWORD", rabbitPass).
		WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("rabbitmq: %w", err)
	}

	// NATS
	fmt.Println("==> NATS JetStream")
	cmd = "helm repo add nats https://nats-io.github.io/k8s/helm/charts --force-update && " +
		"helm upgrade --install nats nats/nats --namespace messaging " +
		"--set nats.jetstream.enabled=true --set cluster.enabled=true --set cluster.replicas=3 --wait --timeout 5m"
	if action == "UNINSTALL" {
		cmd = "helm uninstall nats -n messaging 2>/dev/null || true"
	}
	if _, err := helmBase().WithExec([]string{"sh", "-c", cmd}).Stdout(ctx); err != nil {
		return fmt.Errorf("nats: %w", err)
	}

	// Create Kafka topics
	if getEnv("CREATE_TOPICS", "false") == "true" {
		fmt.Println("==> Creating Kafka topics")
		topics := []string{
			"identity.user.registered", "identity.user.deleted",
			"commerce.order.placed", "commerce.order.cancelled", "commerce.order.fulfilled",
			"commerce.payment.processed", "commerce.payment.failed", "commerce.cart.abandoned",
			"supplychain.shipment.created", "supplychain.shipment.updated",
			"supplychain.inventory.low", "supplychain.inventory.restocked",
			"notification.email.requested", "notification.sms.requested", "notification.push.requested",
			"analytics.page.viewed", "analytics.product.clicked", "analytics.search.performed",
			"security.fraud.detected", "security.login.failed",
		}
		createCmds := []string{
			`KAFKA_POD=$(kubectl get pods -n messaging -l app.kubernetes.io/name=kafka -o jsonpath='{.items[0].metadata.name}')`,
		}
		for _, t := range topics {
			createCmds = append(createCmds,
				fmt.Sprintf(`kubectl exec -n messaging "$KAFKA_POD" -- kafka-topics.sh --create --if-not-exists `+
					`--bootstrap-server kafka.messaging.svc:9092 --topic %s --partitions 12 --replication-factor 3 || true`, t))
		}
		script := strings.Join(createCmds, "\n")
		if _, err := client.Container().
			From("bitnami/kubectl:latest").
			WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"}).
			WithExec([]string{"sh", "-c", script}).
			Stdout(ctx); err != nil {
			fmt.Println("Topic creation: some errors (non-blocking):", err)
		}
	}

	fmt.Println("Messaging pipeline complete")
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
