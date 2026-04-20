def call() {
    sh '''
        echo "=== Install Notation ==="

        if command -v notation >/dev/null 2>&1; then
            echo "  Notation already installed: $(notation version 2>/dev/null | head -1)"
            return 0
        fi

        NOTATION_VERSION=$(curl -sf https://api.github.com/repos/notaryproject/notation/releases/latest \
            | python3 -c "import json,sys; print(json.load(sys.stdin)['tag_name'].lstrip('v'))" 2>/dev/null || echo "1.2.0")

        curl -sfL "https://github.com/notaryproject/notation/releases/download/v${NOTATION_VERSION}/notation_${NOTATION_VERSION}_linux_amd64.tar.gz" \
            -o /tmp/notation.tar.gz
        tar -xzf /tmp/notation.tar.gz -C /tmp notation
        sudo install -o root -g root -m 0755 /tmp/notation /usr/local/bin/notation
        rm -f /tmp/notation.tar.gz /tmp/notation

        notation version
        echo "Notation ${NOTATION_VERSION} installed."
    '''
}
return this
