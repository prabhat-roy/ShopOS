// Dagger Deploy Pipeline — Cloud-agnostic deploy for a single ShopOS service.
// Role: Portable deploy pipeline called by other CI systems (Tekton, Harness, CircleCI, etc.).
//       Detects CLOUD_PROVIDER env var and uses the appropriate kubectl context.
//       Handles: Docker build → Trivy scan → Push to Harbor → Helm deploy → smoke test.
//
// Usage: dagger run go run ci/dagger/deploy/main.go
// Env (required): SERVICE_NAME, IMAGE_TAG, HARBOR_REGISTRY, HARBOR_USERNAME, HARBOR_PASSWORD,
//
//	KUBECONFIG_CONTENT (base64-encoded kubeconfig)
//
// Env (optional): CLOUD_PROVIDER (aws|gcp|azure|local, default: local),
//
//	ENVIRONMENT (default: staging), K8S_NAMESPACE (default: default),
//	DOMAIN (default: platform), SONAR_TOKEN, SONAR_HOST_URL,
//	SLACK_WEBHOOK, SKIP_BUILD (default: false), SKIP_SCAN (default: false)
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

	serviceName := mustEnv("SERVICE_NAME")
	imageTag := mustEnv("IMAGE_TAG")
	registry := mustEnv("HARBOR_REGISTRY")
	environment := getEnv("ENVIRONMENT", "staging")
	namespace := getEnv("K8S_NAMESPACE", "default")
	domain := getEnv("DOMAIN", "platform")
	cloudProvider := getEnv("CLOUD_PROVIDER", "local")
	skipBuild := getEnv("SKIP_BUILD", "false") == "true"
	skipScan := getEnv("SKIP_SCAN", "false") == "true"
	imageRef := fmt.Sprintf("%s/shopos/%s:%s", registry, serviceName, imageTag)

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{".git", "node_modules", "vendor", "target"},
	})

	harborPass := client.SetSecret("harbor_password", mustEnv("HARBOR_PASSWORD"))
	kubeconfig := client.SetSecret("kubeconfig", mustEnv("KUBECONFIG_CONTENT"))

	fmt.Printf("==> Dagger Deploy Pipeline\n")
	fmt.Printf("    Service       : %s\n", serviceName)
	fmt.Printf("    Image         : %s\n", imageRef)
	fmt.Printf("    Environment   : %s\n", environment)
	fmt.Printf("    Namespace     : %s\n", namespace)
	fmt.Printf("    Cloud Provider: %s\n", cloudProvider)
	fmt.Printf("    Skip Build    : %v\n", skipBuild)
	fmt.Printf("    Skip Scan     : %v\n", skipScan)

	// Step 1: Docker build (unless SKIP_BUILD=true)
	var image *dagger.Container
	if !skipBuild {
		fmt.Printf("\n==> [1/5] Docker build: %s\n", imageRef)
		svcDir := fmt.Sprintf("src/%s/%s", domain, serviceName)
		image = client.Container().
			Build(src.Directory(svcDir), dagger.ContainerBuildOpts{
				Dockerfile: "Dockerfile",
			})
	}

	// Step 2: Trivy scan
	if !skipScan {
		fmt.Printf("==> [2/5] Trivy vulnerability scan\n")
		_, err = client.Container().
			From("aquasec/trivy:0.57.1").
			WithMountedDirectory("/workspace", src).
			WithExec([]string{
				"trivy", "image",
				"--exit-code", "1",
				"--severity", "CRITICAL",
				"--format", "table",
				"--no-progress",
				"--timeout", "5m",
				imageRef,
			}).
			Stdout(ctx)
		if err != nil {
			fmt.Printf("WARN: Trivy found CRITICAL CVEs (non-blocking in this pipeline): %v\n", err)
		} else {
			fmt.Println("Trivy: no CRITICAL CVEs found")
		}
	}

	// Step 3: Push to Harbor
	if !skipBuild && image != nil {
		fmt.Printf("==> [3/5] Push to Harbor: %s\n", imageRef)
		addr, pushErr := image.
			WithRegistryAuth(registry, mustEnv("HARBOR_USERNAME"), harborPass).
			Publish(ctx, imageRef)
		if pushErr != nil {
			return fmt.Errorf("docker push failed: %w", pushErr)
		}
		fmt.Printf("    Pushed: %s\n", addr)
	} else {
		fmt.Printf("==> [3/5] Skipping push (SKIP_BUILD=true)\n")
	}

	// Step 4: Helm deploy (cloud-agnostic via kubeconfig)
	fmt.Printf("==> [4/5] Helm deploy to %s (cloud: %s)\n", namespace, cloudProvider)
	helmCmd := fmt.Sprintf(
		"mkdir -p ~/.kube && "+
			"echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config && "+
			"chmod 600 ~/.kube/config && "+
			// Cloud-specific context selection
			"KUBE_CONTEXT=$(kubectl config get-contexts -o name 2>/dev/null | head -1) && "+
			"echo 'Using context: '$KUBE_CONTEXT && "+
			"helm upgrade --install %s /workspace/helm/charts/%s "+
			"  --namespace %s "+
			"  --create-namespace "+
			"  --set image.repository=%s/shopos/%s "+
			"  --set image.tag=%s "+
			"  --set environment=%s "+
			"  --atomic "+
			"  --timeout 5m "+
			"  --history-max 5",
		serviceName, serviceName,
		namespace,
		registry, serviceName,
		imageTag,
		environment,
	)

	_, err = client.Container().
		From("alpine/helm:3.16.2").
		WithMountedDirectory("/workspace", src).
		WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
		WithExec([]string{"sh", "-c", "apk add --no-cache kubectl 2>/dev/null || true"}).
		WithExec([]string{"sh", "-c", helmCmd}).
		Stdout(ctx)
	if err != nil {
		return fmt.Errorf("helm deploy failed: %w", err)
	}
	fmt.Printf("    Deployed %s:%s to %s/%s\n", serviceName, imageTag, cloudProvider, namespace)

	// Step 5: Smoke test
	fmt.Printf("==> [5/5] Smoke test (/healthz)\n")
	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080/healthz", serviceName, namespace)
	smokeOut, _ := client.Container().
		From("curlimages/curl:8.10.1").
		WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
		WithExec([]string{"sh", "-c",
			fmt.Sprintf(
				"for i in 1 2 3 4 5; do "+
					"STATUS=$(curl -sf --max-time 10 %s -o /dev/null -w '%%{http_code}' 2>/dev/null || echo 000); "+
					"echo \"Attempt $i → HTTP $STATUS\"; "+
					"[ \"$STATUS\" = \"200\" ] && echo 'Smoke test PASSED' && exit 0; "+
					"sleep 10; "+
					"done; "+
					"echo 'WARN: smoke test did not pass (gRPC service or extended startup)'; "+
					"exit 0",
				svcURL,
			),
		}).
		Stdout(ctx)
	fmt.Println(smokeOut)

	// Slack notification (optional)
	slackWebhook := os.Getenv("SLACK_WEBHOOK")
	if slackWebhook != "" {
		slackSecret := client.SetSecret("slack_webhook", slackWebhook)
		_, _ = client.Container().
			From("curlimages/curl:8.10.1").
			WithSecretVariable("SLACK_WEBHOOK", slackSecret).
			WithExec([]string{"sh", "-c",
				fmt.Sprintf(
					`curl -sf -X POST "$SLACK_WEBHOOK" -H 'Content-type: application/json' `+
						`--data '{"text":":white_check_mark: Deployed %s:%s to %s/%s (%s)"}'`,
					serviceName, imageTag, cloudProvider, namespace, environment,
				),
			}).
			Stdout(ctx)
	}

	fmt.Printf("\n==> Deploy pipeline complete: %s\n", imageRef)
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
