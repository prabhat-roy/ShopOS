def call() {
    sh '''
        echo "=== Install Skopeo ==="

        if command -v skopeo >/dev/null 2>&1; then
            echo "  Skopeo already installed: $(skopeo --version)"
            return 0
        fi

        # Install via apt on Debian/Ubuntu
        if command -v apt-get >/dev/null 2>&1; then
            sudo apt-get install -y skopeo 2>/dev/null || true
        elif command -v yum >/dev/null 2>&1; then
            sudo yum install -y skopeo 2>/dev/null || true
        else
            # Binary install fallback
            curl -sfL "https://github.com/lework/skopeo-binary/releases/latest/download/skopeo-linux-amd64" \
                -o /tmp/skopeo 2>/dev/null || true
            sudo install -o root -g root -m 0755 /tmp/skopeo /usr/local/bin/skopeo
            rm -f /tmp/skopeo
        fi

        skopeo --version || echo "  Warning: skopeo install may have failed"
        echo "Skopeo installation complete."
    '''
}
return this
