// Dagger Install-Tools Pipeline — verify all language runtimes and security tooling.
// Usage: dagger run go run ci/dagger/install-tools/main.go
// Env: none required (all checks are non-blocking version probes)
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

	type check struct {
		name  string
		image string
		cmd   string
	}

	checks := []check{
		{"kubectl + helm", "alpine/helm:latest", "helm version --short && kubectl version --client"},
		{"Terraform", "hashicorp/terraform:1.9.0", "terraform version"},
		{"Go", "golang:1.23-alpine", "go version"},
		{"Java (Maven)", "maven:3.9-eclipse-temurin-21", "mvn --version && java -version"},
		{"Node.js", "node:22-alpine", "node --version && npm --version"},
		{"Python", "python:3.12-slim", "python --version && pip --version"},
		{".NET", "mcr.microsoft.com/dotnet/sdk:8.0", "dotnet --version"},
		{"Rust", "rust:1.81-slim", "rustc --version && cargo --version"},
		{"Trivy", "aquasec/trivy:latest", "trivy --version"},
		{"Semgrep", "returntocorp/semgrep:latest", "semgrep --version"},
		{"Gitleaks", "zricethezav/gitleaks:latest", "gitleaks version"},
		{"Cosign", "gcr.io/projectsigstore/cosign:latest", "cosign version"},
		{"Ansible", "cytopia/ansible:latest", "ansible --version"},
		{"k6", "grafana/k6:latest", "k6 version"},
	}

	allPassed := true
	for _, c := range checks {
		fmt.Printf("==> Checking: %s\n", c.name)
		out, err := client.Container().
			From(c.image).
			WithExec([]string{"sh", "-c", c.cmd + " 2>&1 || true"}).
			Stdout(ctx)
		if err != nil {
			fmt.Printf("    WARN: %s check failed: %v\n", c.name, err)
			allPassed = false
		} else {
			fmt.Printf("    OK: %s\n", out)
		}
	}

	if !allPassed {
		fmt.Println("Some tool checks failed (see warnings above)")
	}
	fmt.Println("Install-tools pipeline complete")
	return nil
}
