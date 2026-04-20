def call() {
    sh '''
        echo "=== Install gcloud CLI ==="

        if command -v gcloud >/dev/null 2>&1; then
            echo "  gcloud already installed: $(gcloud --version | head -1)"
            return 0
        fi

        curl -sf "https://sdk.cloud.google.com" -o /tmp/install-gcloud.sh
        sudo bash /tmp/install-gcloud.sh --disable-prompts --install-dir=/usr/local 2>/dev/null || true
        rm -f /tmp/install-gcloud.sh

        # Add to PATH for current session
        export PATH="$PATH:/usr/local/google-cloud-sdk/bin"
        sudo ln -sf /usr/local/google-cloud-sdk/bin/gcloud /usr/local/bin/gcloud 2>/dev/null || true
        sudo ln -sf /usr/local/google-cloud-sdk/bin/gsutil /usr/local/bin/gsutil 2>/dev/null || true

        gcloud --version | head -1 || echo "  Warning: gcloud install may have failed"

        # Install gke-gcloud-auth-plugin (required for kubectl to auth with GKE since k8s 1.26)
        gcloud components install gke-gcloud-auth-plugin --quiet 2>/dev/null || \
            sudo apt-get install -y google-cloud-cli-gke-gcloud-auth-plugin 2>/dev/null || \
            sudo yum install -y google-cloud-cli-gke-gcloud-auth-plugin 2>/dev/null || \
            echo "  Warning: gke-gcloud-auth-plugin install failed — kubectl will not work with GKE"

        echo "gcloud CLI installation complete."
    '''
}
return this
