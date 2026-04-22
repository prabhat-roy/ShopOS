// Dagger Docker-Build Pipeline — build, scan (Trivy), and push all ShopOS images to Harbor.
// Role: Portable image build pipeline. Runs Trivy vulnerability scan after each build.
//       Can be called from any CI system: Jenkins, CircleCI, Travis, Harness, etc.
//
// Usage: dagger run go run ci/dagger/docker-build/main.go
// Env (required): HARBOR_REGISTRY, HARBOR_USERNAME, HARBOR_PASSWORD
// Env (optional): SERVICE_NAME, DOMAIN, IMAGE_TAG, REGISTRY_PROJECT, PUSH_IMAGE, SCAN_IMAGE,
//
//	SCAN_SEVERITY (default: HIGH,CRITICAL), FAIL_ON_CRITICAL (default: false)
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

	registry := mustEnv("HARBOR_REGISTRY")
	registryProject := getEnv("REGISTRY_PROJECT", "shopos")
	serviceName := os.Getenv("SERVICE_NAME")
	domain := os.Getenv("DOMAIN")
	imageTag := getEnv("IMAGE_TAG", "dev-dagger")
	pushImage := getEnv("PUSH_IMAGE", "true") == "true"
	scanImage := getEnv("SCAN_IMAGE", "true") == "true"
	scanSeverity := getEnv("SCAN_SEVERITY", "HIGH,CRITICAL")
	failOnCritical := getEnv("FAIL_ON_CRITICAL", "false") == "true"

	harborPass := client.SetSecret("harbor_password", mustEnv("HARBOR_PASSWORD"))

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{".git", "node_modules", "vendor", "target"},
	})

	// Resolve which services to build
	targets, err := resolveTargets(serviceName, domain)
	if err != nil {
		return fmt.Errorf("resolving targets: %w", err)
	}

	fmt.Printf("==> Docker-Build pipeline (with Trivy scan)\n")
	fmt.Printf("    Registry  : %s/%s\n", registry, registryProject)
	fmt.Printf("    Tag       : %s\n", imageTag)
	fmt.Printf("    Targets   : %d services\n", len(targets))
	fmt.Printf("    Push      : %v\n", pushImage)
	fmt.Printf("    Scan      : %v (severity: %s)\n", scanImage, scanSeverity)
	fmt.Printf("    FailCrit  : %v\n", failOnCritical)

	var buildErrors []string

	for _, t := range targets {
		imageRef := fmt.Sprintf("%s/%s/%s:%s", registry, registryProject, t.svc, imageTag)
		svcDir := fmt.Sprintf("src/%s/%s", t.domain, t.svc)

		// Step 1: Docker build
		fmt.Printf("\n==> [%s/%s] Building %s\n", t.domain, t.svc, imageRef)
		image := client.Container().
			Build(src.Directory(svcDir), dagger.ContainerBuildOpts{
				Dockerfile: "Dockerfile",
			})

		// Step 2: Trivy vulnerability scan
		if scanImage {
			fmt.Printf("==> [%s] Trivy scan (severity: %s)\n", t.svc, scanSeverity)
			exitCode := "0"
			if failOnCritical {
				exitCode = "1"
			}
			trivyOut, trivyErr := client.Container().
				From("aquasec/trivy:0.57.1").
				WithMountedDirectory("/workspace", src).
				WithExec([]string{
					"trivy", "image",
					"--exit-code", exitCode,
					"--severity", scanSeverity,
					"--format", "table",
					"--no-progress",
					"--timeout", "5m",
					imageRef,
				}).
				Stdout(ctx)
			if trivyErr != nil {
				msg := fmt.Sprintf("[%s] Trivy scan: %v", t.svc, trivyErr)
				if failOnCritical {
					buildErrors = append(buildErrors, msg)
					fmt.Printf("FAIL: %s\n", msg)
					continue
				}
				fmt.Printf("WARN: %s (non-blocking)\n", msg)
			} else {
				fmt.Println(trivyOut)
			}

			// Generate SBOM (always non-blocking)
			fmt.Printf("==> [%s] Generating SBOM (CycloneDX)\n", t.svc)
			_, _ = client.Container().
				From("aquasec/trivy:0.57.1").
				WithExec([]string{
					"trivy", "image",
					"--format", "cyclonedx",
					"--output", fmt.Sprintf("/tmp/sbom-%s.json", t.svc),
					"--no-progress",
					imageRef,
				}).
				Stdout(ctx)
		}

		// Step 3: Push to Harbor
		if pushImage {
			fmt.Printf("==> [%s] Pushing %s\n", t.svc, imageRef)
			addr, pushErr := image.
				WithRegistryAuth(registry, mustEnv("HARBOR_USERNAME"), harborPass).
				Publish(ctx, imageRef)
			if pushErr != nil {
				msg := fmt.Sprintf("[%s] push failed: %v", t.svc, pushErr)
				buildErrors = append(buildErrors, msg)
				fmt.Printf("FAIL: %s\n", msg)
				continue
			}
			fmt.Printf("    Pushed: %s\n", addr)

			// Also push with :latest tag
			latestRef := fmt.Sprintf("%s/%s/%s:latest", registry, registryProject, t.svc)
			_, _ = image.
				WithRegistryAuth(registry, mustEnv("HARBOR_USERNAME"), harborPass).
				Publish(ctx, latestRef)
		}
	}

	if len(buildErrors) > 0 {
		fmt.Printf("\n==> Build errors (%d):\n", len(buildErrors))
		for _, e := range buildErrors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("%d services failed to build/push/scan", len(buildErrors))
	}

	fmt.Printf("\n==> Docker-Build pipeline complete. %d services processed.\n", len(targets))
	return nil
}

// resolveTargets returns the list of (domain, svc) pairs to build.
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
				// Only include if Dockerfile exists
				if _, err := os.Stat(filepath.Join("src", d, e.Name(), "Dockerfile")); err == nil {
					targets = append(targets, buildTarget{domain: d, svc: e.Name()})
				}
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
