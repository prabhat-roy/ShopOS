def call() {
    sh '''
        echo "=== Install Cosign ==="

        if command -v cosign >/dev/null 2>&1; then
            echo "  Cosign already installed: $(cosign version 2>/dev/null | head -1)"
            return 0
        fi

        COSIGN_VERSION=$(curl -sf https://api.github.com/repos/sigstore/cosign/releases/latest \
            | python3 -c "import json,sys; print(json.load(sys.stdin)['tag_name'])" 2>/dev/null || echo "v2.4.0")

        curl -sfL "https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign-linux-amd64" \
            -o /tmp/cosign
        sudo install -o root -g root -m 0755 /tmp/cosign /usr/local/bin/cosign
        rm -f /tmp/cosign

        cosign version
        echo "Cosign ${COSIGN_VERSION} installed."
    '''
}
return this
