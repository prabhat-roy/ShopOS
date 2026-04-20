def call() {
    sh """
        ANCHORE_URL=\$(grep '^ANCHORE_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Anchore at \${ANCHORE_URL}..."
        until curl -sf -u admin:foobar "\${ANCHORE_URL}/v1/status" > /dev/null 2>&1; do sleep 10; done

        # Add Docker Hub registry (unauthenticated public access)
        curl -sf -u admin:foobar -X POST "\${ANCHORE_URL}/v1/registries" \
            -H "Content-Type: application/json" \
            -d '{"registry":"docker.io","registry_type":"docker_v2","registry_user":"","registry_pass":"","registry_verify":false}' || true

        # Add ghcr.io registry
        curl -sf -u admin:foobar -X POST "\${ANCHORE_URL}/v1/registries" \
            -H "Content-Type: application/json" \
            -d '{"registry":"ghcr.io","registry_type":"docker_v2","registry_user":"","registry_pass":"","registry_verify":false}' || true

        # Set policy to activate default bundle
        POLICY_ID=\$(curl -sf -u admin:foobar "\${ANCHORE_URL}/v1/policies" \
            | grep -o '"policyId":"[^"]*"' | head -1 | cut -d: -f2 | tr -d '"')
        curl -sf -u admin:foobar -X PUT "\${ANCHORE_URL}/v1/policies/\${POLICY_ID}" \
            -H "Content-Type: application/json" \
            -d '{"active":true}' || true
    """
    echo 'anchore configured — registries added, default policy activated'
}
return this
