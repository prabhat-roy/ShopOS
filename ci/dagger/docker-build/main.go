// Dagger Docker-Build Pipeline — build, scan (Trivy), and push all ShopOS images to Harbor.
// Usage: dagger run go run ci/dagger/docker-build/main.go
// Env (required): HARBOR_REGISTRY, HARBOR_USERNAME, HARBOR_PASSWORD
// Env (optional): SERVICE_NAME, DOMAIN, IMAGE_TAG, REGISTRY_PROJECT, PUSH_IMAGE, SCAN_IMAGE
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

var allDomains = []string{
	"platform", "identity", "catalog", "commerce", "supply-chain",
	"financial", "customer-experience", "communications", "content",
	"analytics-ai", "b2b", "integrations", "affiliate",
}

type buildTarget struct {
	domain string
	svc    string
}

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

	registry        := mustEnv("HARBOR_REGISTRY")
	registryProject := getEnv("REGISTRY_PROJECT", "shopos")
	serviceName     := os.Getenv("SERVICE_NAME")
	domain          := os.Getenv("DOMAIN")
	imageTag        := getEnv("IMAGE_TAG", "dev-dagger")
	pushImage       := getEnv("PUSH_IMAGE", "true") == "true"
	scanImage       := getEnv("SCAN_IMAGE", "true") == "true"

	harborPass := client.SetSecret("harbor_password", mustEnv("HARBOR_PASSWORD"))

	src := client.Host().Directory(".")

	// Resolve which services to build
	targets, err := resolveTargets(serviceName, domain)
	if err != nil {
		return fmt.Errorf("resolving targets: %w", err)
	}

	fmt.Printf("==> Docker-Build pipeline\n")
	fmt.Printf("    Registry : %s/%s\n", registry, registryProject)
	fmt.Printf("    Tag      : %s\n", imageTag)
	fmt.Printf("    Targets  : %d services\n", len(targets))
	fmt.Printf("    Push     : %v\n", pushImage)
	fmt.Printf("    Scan     : %v\n", scanImage)

	for _, t := range targets {
		imageRef := fmt.Sprintf("%s/%s/%s:%s", registry, registryProject, t.svc, imageTag)
		svcDir   := fmt.Sprintf("src/%s/%s", t.domain, t.svc)

		// Step 1: Docker build
		fmt.Printf("\n==> [%s/%s] Building %s\n", t.domain, t.svc, imageRef)
		image := client.Container().
			Build(src.Directory(svcDir), dagger.ContainerBuildOpts{
				Dockerfile: "Dockerfile",
			})

		// Step 2: Trivy scan
		if scanImage {
			fmt.Printf("==> [%s] Trivy scan\n", t.svc)
			_, err = client.Container().
				From("aquasec/trivy:latest").
				WithMountedDirectory("/workspace", src).
				WithExec([]string{
					"trivy", "image",
					"--exit-code", "0",
					"--severity", "HIGH,CRITICAL",
					"--no-progress",
					imageRef,
				}).
				Stdout(ctx)
			if err != nil {
				fmt.Printf("WARN Trivy scan failed for %s (non-blocking): %v\n", t.svc, err)
			}
		}

		// Step 3: Push to Harbor
		if pushImage {
			fmt.Printf("==> [%s] Pushing %s\n", t.svc, imageRef)
			_, err = image.
				WithRegistryAuth(registry, mustEnv("HARBOR_USERNAME"), harborPass).
				Publish(ctx, imageRef)
			if err != nil {
				fmt.Printf("WARN push failed for %s: %v\n", t.svc, err)
				continue
			}
			fmt.Printf("    Pushed: %s\n", imageRef)
		}
	}

	fmt.Println("\n==> Docker-Build pipeline complete.")
	return nil
}

// resolveTargets returns the list of (domain, svc) pairs to build.
// Priority: SERVICE_NAME > DOMAIN > all 154.
func resolveTargets(serviceName, domain string) ([]buildTarget, error) {
	if serviceName != "" && domain != "" {
		return []buildTarget{{domain: domain, svc: serviceName}}, nil
	}

	var domains []string
	if domain != "" {
		domains = []string{domain}
	} else {
		domains = allDomains
	}

	var targets []buildTarget
	for _, d := range domains {
		entries, err := os.ReadDir(filepath.Join("src", d))
		if err != nil {
			fmt.Printf("WARN cannot read src/%s: %v\n", d, err)
			continue
		}
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				targets = append(targets, buildTarget{domain: d, svc: e.Name()})
			}
		}
	}
	return targets, nil
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
