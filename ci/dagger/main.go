// Dagger CI pipeline for ShopOS — runs alongside Jenkins and Drone CI.
// Usage: dagger run go run ci/dagger/main.go <pipeline>
// Pipelines: build, test, lint, publish
package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: main.go <build|test|lint|publish>")
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	switch os.Args[1] {
	case "build":
		build(ctx, client)
	case "test":
		test(ctx, client)
	case "lint":
		lint(ctx, client)
	case "publish":
		publish(ctx, client)
	default:
		fmt.Fprintf(os.Stderr, "unknown pipeline: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func build(ctx context.Context, client *dagger.Client) {
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{".git", "node_modules", "vendor"},
	})

	// Example: build Go services
	golang := client.Container().
		From("golang:1.24-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"sh", "-c", "find src -name 'main.go' -path '*/platform/*' | head -5 | xargs -I{} dirname {} | xargs -I{} go build ./{}"})

	if _, err := golang.Sync(ctx); err != nil {
		panic(err)
	}
	fmt.Println("build: OK")
}

func test(ctx context.Context, client *dagger.Client) {
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{".git", "node_modules"},
	})

	result, err := client.Container().
		From("golang:1.24-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "./..."}).
		Stdout(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func lint(ctx context.Context, client *dagger.Client) {
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{".git"},
	})

	result, err := client.Container().
		From("golangci/golangci-lint:v1.61").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"golangci-lint", "run", "--timeout=5m", "./..."}).
		Stdout(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func publish(ctx context.Context, client *dagger.Client) {
	registry := os.Getenv("REGISTRY")
	tag := os.Getenv("IMAGE_TAG")
	if registry == "" {
		registry = "harbor.shopos.internal"
	}
	if tag == "" {
		tag = "latest"
	}

	src := client.Host().Directory("src/platform/api-gateway")

	image := client.Container().
		Build(src, dagger.ContainerBuildOpts{Dockerfile: "Dockerfile"}).
		WithRegistryAuth(registry, os.Getenv("REGISTRY_USER"), client.SetSecret("registry-pass", os.Getenv("REGISTRY_PASSWORD")))

	addr, err := image.Publish(ctx, fmt.Sprintf("%s/shopos/api-gateway:%s", registry, tag))
	if err != nil {
		panic(err)
	}
	fmt.Println("published:", addr)
}
