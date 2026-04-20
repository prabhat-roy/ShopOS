// Dagger Ansible Pipeline — lint and run Ansible playbooks for K8s node bootstrapping.
// Usage: dagger run go run ci/dagger/ansible/main.go
// Env: ANSIBLE_TARGET_HOSTS, SSH_PRIVATE_KEY, ANSIBLE_USER (default: ubuntu),
//      PLAYBOOK (default: site.yml), ANSIBLE_TAGS (optional)
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

	src := client.Host().Directory(".")
	sshKey := client.SetSecret("ssh_key", mustEnv("SSH_PRIVATE_KEY"))
	targetHosts := mustEnv("ANSIBLE_TARGET_HOSTS")
	ansibleUser := getEnv("ANSIBLE_USER", "ubuntu")
	playbook := getEnv("PLAYBOOK", "site.yml")
	tags := getEnv("ANSIBLE_TAGS", "")

	ansibleBase := client.Container().
		From("cytopia/ansible:latest").
		WithMountedDirectory("/workspace", src).
		WithWorkdir("/workspace/infra/ansible")

	// Step 1: Lint
	fmt.Println("==> Ansible lint")
	if _, err := ansibleBase.
		WithExec([]string{"sh", "-c", "ansible-lint *.yml 2>/dev/null || true"}).
		Stdout(ctx); err != nil {
		fmt.Println("Ansible lint: warnings (non-blocking):", err)
	}

	// Step 2: Syntax check
	fmt.Println("==> Ansible syntax check")
	if _, err := ansibleBase.
		WithExec([]string{"ansible-playbook", "--syntax-check", playbook}).
		Stdout(ctx); err != nil {
		return fmt.Errorf("ansible syntax check failed: %w", err)
	}

	// Step 3: Run playbook
	fmt.Println("==> Ansible playbook:", playbook, "→", targetHosts)
	playbookCmd := "mkdir -p ~/.ssh && " +
		"echo \"$SSH_PRIVATE_KEY\" > ~/.ssh/id_rsa && " +
		"chmod 600 ~/.ssh/id_rsa && " +
		"ansible-playbook " + playbook +
		" -i " + targetHosts + "," +
		" -u " + ansibleUser +
		" --private-key ~/.ssh/id_rsa" +
		" --ssh-extra-args='-o StrictHostKeyChecking=no'"
	if tags != "" {
		playbookCmd += " --tags=" + tags
	}
	playbookCmd += " && rm -f ~/.ssh/id_rsa"

	if _, err := ansibleBase.
		WithSecretVariable("SSH_PRIVATE_KEY", sshKey).
		WithExec([]string{"sh", "-c", playbookCmd}).
		Stdout(ctx); err != nil {
		return fmt.Errorf("ansible playbook failed: %w", err)
	}

	fmt.Println("Ansible pipeline complete")
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
