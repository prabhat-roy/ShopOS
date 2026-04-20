def call() {
    sh """
        VAULT_ADDR=\$(grep '^VAULT_URL=' infra.env | cut -d= -f2)
        export VAULT_ADDR

        echo "Waiting for Vault at \${VAULT_ADDR}..."
        until curl -sf "\${VAULT_ADDR}/v1/sys/health" > /dev/null 2>&1; do sleep 5; done

        # Initialize Vault (1 key share, threshold 1 for simplicity)
        INIT_OUT=\$(curl -sf -X PUT "\${VAULT_ADDR}/v1/sys/init" \
            -H "Content-Type: application/json" \
            -d '{"secret_shares":1,"secret_threshold":1}')
        UNSEAL_KEY=\$(echo "\${INIT_OUT}" | grep -o '"keys":\\["[^"]*"' | grep -o '"[^"]*"' | tail -1 | tr -d '"')
        ROOT_TOKEN=\$(echo "\${INIT_OUT}" | grep -o '"root_token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Unseal
        curl -sf -X PUT "\${VAULT_ADDR}/v1/sys/unseal" \
            -H "Content-Type: application/json" \
            -d "{\\"key\\":\\"\${UNSEAL_KEY}\\"}"

        # Enable KV v2 secrets engine
        curl -sf -X POST "\${VAULT_ADDR}/v1/sys/mounts/secret" \
            -H "X-Vault-Token: \${ROOT_TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"type":"kv","options":{"version":"2"}}' || true

        # Enable PKI secrets engine
        curl -sf -X POST "\${VAULT_ADDR}/v1/sys/mounts/pki" \
            -H "X-Vault-Token: \${ROOT_TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"type":"pki"}' || true

        sed -i '/^VAULT_ROOT_TOKEN=/d' infra.env || true
        sed -i '/^VAULT_UNSEAL_KEY=/d' infra.env || true
        echo "VAULT_ROOT_TOKEN=\${ROOT_TOKEN}" >> infra.env
        echo "VAULT_UNSEAL_KEY=\${UNSEAL_KEY}" >> infra.env
    """
    echo 'vault configured — root token and unseal key written to infra.env'
}
return this
