// Dagger Post-Deploy Pipeline — health checks, load test, DAST, kube-bench, Slack notify.
// Usage: dagger run go run ci/dagger/post-deploy/main.go
// Env: KUBECONFIG_CONTENT, SERVICE_NAME, SERVICE_URL, ENVIRONMENT,
//      SLACK_WEBHOOK, ZAP_TARGET_URL (default: SERVICE_URL)
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
	serviceURL := mustEnv("SERVICE_URL")
	environment := getEnv("ENVIRONMENT", "staging")
	zapTarget := getEnv("ZAP_TARGET_URL", serviceURL)
	kubeconfig := client.SetSecret("kubeconfig", mustEnv("KUBECONFIG_CONTENT"))
	slackWebhook := client.SetSecret("slack_webhook", mustEnv("SLACK_WEBHOOK"))
	src := client.Host().Directory(".")

	status := "passed"
	details := ""

	// Step 1: Health check
	fmt.Println("==> Health check:", serviceURL)
	_, err = client.Container().
		From("curlimages/curl:latest").
		WithExec([]string{"sh", "-c",
			fmt.Sprintf("for i in 1 2 3 4 5; do curl -sf %s/healthz && exit 0; sleep 10; done; exit 1", serviceURL)}).
		Stdout(ctx)
	if err != nil {
		status = "failed"
		details += " health-check-failed"
		fmt.Println("Health check FAILED:", err)
	}

	// Step 2: k6 smoke test
	fmt.Println("==> k6 smoke test")
	_, err = client.Container().
		From("grafana/k6:latest").
		WithMountedDirectory("/workspace", src).
		WithEnvVariable("TARGET_URL", serviceURL).
		WithExec([]string{"k6", "run", "--vus", "5", "--duration", "30s",
			"/workspace/load-testing/k6/smoke-test.js"}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("k6 smoke test: issues (non-blocking):", err)
	}

	// Step 3: ZAP DAST baseline
	fmt.Println("==> ZAP DAST baseline scan:", zapTarget)
	_, err = client.Container().
		From("ghcr.io/zaproxy/zaproxy:stable").
		WithExec([]string{"zap-baseline.py", "-t", zapTarget, "-r", "/zap/wrk/zap-report.html", "--auto"}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("ZAP DAST: alerts found (non-blocking):", err)
	}

	// Step 4: kube-bench CIS audit
	fmt.Println("==> kube-bench CIS benchmark")
	_, err = client.Container().
		From("aquasec/kube-bench:latest").
		WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
		WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"}).
		WithExec([]string{"kube-bench", "--json"}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("kube-bench: findings (non-blocking):", err)
	}

	// Step 5: kubescape scan
	fmt.Println("==> Kubescape security posture")
	_, err = client.Container().
		From("kubescape/kubescape:latest").
		WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
		WithExec([]string{"sh", "-c", "mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config"}).
		WithExec([]string{"kubescape", "scan", "framework", "nsa", "--format", "json"}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("Kubescape: findings (non-blocking):", err)
	}

	// Step 6: Slack notify
	fmt.Println("==> Slack notification")
	emoji := ":white_check_mark:"
	if status == "failed" {
		emoji = ":x:"
	}
	_, err = client.Container().
		From("curlimages/curl:latest").
		WithSecretVariable("SLACK_WEBHOOK", slackWebhook).
		WithExec([]string{"sh", "-c",
			fmt.Sprintf(`curl -sf -X POST "$SLACK_WEBHOOK" -H 'Content-type: application/json' `+
				`--data '{"text":"%s Post-deploy checks %s for %s in %s%s"}'`,
				emoji, status, serviceName, environment, details)}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("Slack notify failed (non-blocking)")
	}

	if status == "failed" {
		return fmt.Errorf("post-deploy pipeline failed:%s", details)
	}

	fmt.Println("Post-deploy pipeline complete")
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
