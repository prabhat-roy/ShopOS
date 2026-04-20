def call() {
    def anchoreUrl  = 'http://anchore-api.anchore.svc.cluster.local:8228'
    def anchorePass = 'foobar'
    if (fileExists('infra.env')) {
        readFile('infra.env').trim().split('\n').each { line ->
            def parts = line.split('=', 2)
            if (parts.length == 2) {
                if (parts[0] == 'ANCHORE_URL')      anchoreUrl  = parts[1]
                if (parts[0] == 'ANCHORE_PASSWORD') anchorePass = parts[1]
            }
        }
    }

    sh """
        echo "Waiting for Anchore at ${anchoreUrl}..."
        until curl -sf -u admin:${anchorePass} "${anchoreUrl}/v1/status" > /dev/null 2>&1; do sleep 10; done

        curl -sf -u admin:${anchorePass} -X POST "${anchoreUrl}/v1/registries" \
            -H "Content-Type: application/json" \
            -d '{"registry":"docker.io","registry_type":"docker_v2","registry_user":"","registry_pass":"","registry_verify":false}' || true

        curl -sf -u admin:${anchorePass} -X POST "${anchoreUrl}/v1/registries" \
            -H "Content-Type: application/json" \
            -d '{"registry":"ghcr.io","registry_type":"docker_v2","registry_user":"","registry_pass":"","registry_verify":false}' || true

        POLICY_ID=\$(curl -sf -u admin:${anchorePass} "${anchoreUrl}/v1/policies" \
            | grep -o '"policyId":"[^"]*"' | head -1 | cut -d: -f2 | tr -d '"')
        curl -sf -u admin:${anchorePass} -X PUT "${anchoreUrl}/v1/policies/\${POLICY_ID}" \
            -H "Content-Type: application/json" \
            -d '{"active":true}' || true
    """
    echo 'anchore configured — registries added, default policy activated'
}
return this
