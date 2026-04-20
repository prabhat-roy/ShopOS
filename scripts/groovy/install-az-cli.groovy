def call() {
    sh '''
        echo "=== Install Azure CLI ==="

        if command -v az >/dev/null 2>&1; then
            echo "  Azure CLI already installed: $(az --version | head -1)"
            return 0
        fi

        # Install via Microsoft's install script
        curl -sfL https://aka.ms/InstallAzureCLIDeb -o /tmp/install-az.sh
        sudo bash /tmp/install-az.sh 2>/dev/null || true
        rm -f /tmp/install-az.sh

        # Fallback: pip install
        if ! command -v az >/dev/null 2>&1; then
            sudo pip3 install azure-cli 2>/dev/null || true
        fi

        az --version | head -1 || echo "  Warning: az install may have failed"
        echo "Azure CLI installation complete."
    '''
}
return this
