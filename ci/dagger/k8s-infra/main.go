// Dagger K8s Infrastructure Pipeline — Terraform for EKS, GKE, AKS.
// Usage: dagger run go run ci/dagger/k8s-infra/main.go
// Env: CLOUD (aws|gcp|azure), ACTION (apply|destroy),
//      AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION,
//      GCP_PROJECT_ID, GCP_SA_KEY,
//      ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_SUBSCRIPTION_ID, ARM_TENANT_ID
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

	cloud := getEnv("CLOUD", "aws")
	action := getEnv("ACTION", "apply")
	src := client.Host().Directory(".")

	tfBase := func(tfDir string) *dagger.Container {
		return client.Container().
			From("hashicorp/terraform:1.9.0").
			WithMountedDirectory("/workspace", src).
			WithWorkdir("/workspace/" + tfDir).
			WithExec([]string{"terraform", "init", "-input=false"})
	}

	switch cloud {
	case "aws":
		fmt.Println("==> Terraform EKS (AWS)")
		awsKey := client.SetSecret("aws_key", mustEnv("AWS_ACCESS_KEY_ID"))
		awsSecret := client.SetSecret("aws_secret", mustEnv("AWS_SECRET_ACCESS_KEY"))
		cmd := []string{"terraform", action, "-auto-approve", "-input=false"}
		if action == "destroy" {
			cmd = []string{"terraform", "destroy", "-auto-approve", "-input=false"}
		}
		if _, err := tfBase("infra/terraform/aws/eks").
			WithSecretVariable("AWS_ACCESS_KEY_ID", awsKey).
			WithSecretVariable("AWS_SECRET_ACCESS_KEY", awsSecret).
			WithEnvVariable("AWS_REGION", getEnv("AWS_REGION", "us-east-1")).
			WithExec(cmd).Stdout(ctx); err != nil {
			return fmt.Errorf("eks terraform: %w", err)
		}

	case "gcp":
		fmt.Println("==> Terraform GKE (GCP)")
		gcpKey := client.SetSecret("gcp_sa_key", mustEnv("GCP_SA_KEY"))
		cmd := []string{"terraform", action, "-auto-approve", "-input=false"}
		if action == "destroy" {
			cmd = []string{"terraform", "destroy", "-auto-approve", "-input=false"}
		}
		if _, err := tfBase("infra/terraform/gcp/gke").
			WithSecretVariable("GOOGLE_CREDENTIALS", gcpKey).
			WithEnvVariable("GOOGLE_PROJECT", mustEnv("GCP_PROJECT_ID")).
			WithExec(cmd).Stdout(ctx); err != nil {
			return fmt.Errorf("gke terraform: %w", err)
		}

	case "azure":
		fmt.Println("==> Terraform AKS (Azure)")
		clientSecret := client.SetSecret("arm_client_secret", mustEnv("ARM_CLIENT_SECRET"))
		cmd := []string{"terraform", action, "-auto-approve", "-input=false"}
		if action == "destroy" {
			cmd = []string{"terraform", "destroy", "-auto-approve", "-input=false"}
		}
		if _, err := tfBase("infra/terraform/azure/aks").
			WithSecretVariable("ARM_CLIENT_SECRET", clientSecret).
			WithEnvVariable("ARM_CLIENT_ID", mustEnv("ARM_CLIENT_ID")).
			WithEnvVariable("ARM_SUBSCRIPTION_ID", mustEnv("ARM_SUBSCRIPTION_ID")).
			WithEnvVariable("ARM_TENANT_ID", mustEnv("ARM_TENANT_ID")).
			WithExec(cmd).Stdout(ctx); err != nil {
			return fmt.Errorf("aks terraform: %w", err)
		}

	default:
		return fmt.Errorf("unknown CLOUD: %s (valid: aws, gcp, azure)", cloud)
	}

	fmt.Println("K8s infra pipeline complete")
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
