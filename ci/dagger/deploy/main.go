// Dagger Deploy Pipeline — build, scan, push, sign, deploy for a single service.
// Usage: dagger run go run ci/dagger/deploy/main.go
// Env: SERVICE_NAME, IMAGE_TAG, ENVIRONMENT, K8S_NAMESPACE,
//      HARBOR_REGISTRY, HARBOR_USERNAME, HARBOR_PASSWORD,
//      SONAR_TOKEN, SONAR_HOST_URL, KUBECONFIG_CONTENT, SLACK_WEBHOOK
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
	imageRef := fmt.Sprintf("%s/shopos/%s:%s", registry, serviceName, imageTag)

	src := client.Host().Directory(".")

	harborPass := client.SetSecret("harbor_password", mustEnv("HARBOR_PASSWORD"))
	sonarToken := client.SetSecret("sonar_token", mustEnv("SONAR_TOKEN"))
	kubeconfig := client.SetSecret("kubeconfig", mustEnv("KUBECONFIG_CONTENT"))
	slackWebhook := client.SetSecret("slack_webhook", mustEnv("SLACK_WEBHOOK"))

	// Step 1: Secret scan
	fmt.Println("==> Secret scan (Gitleaks)")
	_, err = client.Container().
		From("zricethezav/gitleaks:latest").
		WithMountedDirectory("/repo", src).
		WithExec([]string{"detect", "--source", "/repo", "--exit-code", "1"}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("Gitleaks: warnings found (non-blocking)")
	}

	// Step 2: SAST (Semgrep)
	fmt.Println("==> SAST (Semgrep)")
	_, err = client.Container().
		From("returntocorp/semgrep:latest").
		WithMountedDirectory("/src", src).
		WithExec([]string{"semgrep", "--config=auto", "--json",
			"--output=/tmp/semgrep.json", "/src/src/" + serviceName}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("Semgrep: issues found (non-blocking)")
	}

	// Step 3: SonarQube scan
	fmt.Println("==> SonarQube scan")
	_, err = client.Container().
		From("sonarsource/sonar-scanner-cli:latest").
		WithMountedDirectory("/src", src).
		WithSecretVariable("SONAR_TOKEN", sonarToken).
		WithEnvVariable("SONAR_HOST_URL", mustEnv("SONAR_HOST_URL")).
		WithWorkdir("/src/src/" + serviceName).
		WithExec([]string{"sh", "-c",
			"sonar-scanner -Dsonar.projectKey=" + serviceName +
				" -Dsonar.host.url=$SONAR_HOST_URL -Dsonar.login=$SONAR_TOKEN || true"}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("SonarQube: scan issues (non-blocking)")
	}

	// Step 4: Docker build
	fmt.Println("==> Docker build:", imageRef)
	image := client.Container().
		From("ubuntu:24.04").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src/src/" + serviceName).
		Build(src.Directory("src/"+serviceName), dagger.ContainerBuildOpts{
			Dockerfile: "Dockerfile",
		})

	// Step 5: Trivy image scan
	fmt.Println("==> Trivy image scan")
	imageExported := image.
		WithExec([]string{"echo", "image-built"})
	_, err = client.Container().
		From("aquasec/trivy:latest").
		WithMountedDirectory("/workspace", src).
		WithExec([]string{"trivy", "image", "--exit-code", "1",
			"--severity", "CRITICAL", "--no-progress", imageRef}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("Trivy: critical CVEs found (non-blocking)")
	}
	_ = imageExported

	// Step 6: Push to Harbor
	fmt.Println("==> Push to Harbor:", imageRef)
	_, err = image.
		WithRegistryAuth(registry, mustEnv("HARBOR_USERNAME"), harborPass).
		Publish(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	// Step 7: Cosign sign
	fmt.Println("==> Cosign keyless sign:", imageRef)
	_, err = client.Container().
		From("gcr.io/projectsigstore/cosign:latest").
		WithEnvVariable("COSIGN_EXPERIMENTAL", "1").
		WithExec([]string{"cosign", "sign", "--yes", imageRef}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("Cosign sign failed (non-blocking):", err)
	}

	// Step 8: Helm deploy
	fmt.Println("==> Helm deploy:", serviceName, "→", namespace)
	_, err = client.Container().
		From("alpine/helm:latest").
		WithMountedDirectory("/workspace", src).
		WithSecretVariable("KUBECONFIG_CONTENT", kubeconfig).
		WithExec([]string{"sh", "-c",
			"mkdir -p ~/.kube && echo $KUBECONFIG_CONTENT | base64 -d > ~/.kube/config && " +
				"helm upgrade --install " + serviceName +
				" /workspace/helm/charts/" + serviceName +
				" --namespace " + namespace + " --create-namespace" +
				" --set image.repository=" + registry + "/shopos/" + serviceName +
				" --set image.tag=" + imageTag +
				" --set environment=" + environment +
				" --wait --timeout 5m"}).
		Stdout(ctx)
	if err != nil {
		return fmt.Errorf("helm deploy failed: %w", err)
	}

	// Step 9: Slack notification
	_, err = client.Container().
		From("curlimages/curl:latest").
		WithSecretVariable("SLACK_WEBHOOK", slackWebhook).
		WithExec([]string{"sh", "-c",
			`curl -sf -X POST "$SLACK_WEBHOOK" -H 'Content-type: application/json' ` +
				`--data '{"text":":white_check_mark: Deployed ` + serviceName + `:` + imageTag + ` to ` + environment + `"}'`}).
		Stdout(ctx)
	if err != nil {
		fmt.Println("Slack notify failed (non-blocking)")
	}

	fmt.Println("Deploy pipeline complete:", imageRef)
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
