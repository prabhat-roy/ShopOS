def call() {
    sh '''
        echo "=== Install gke-gcloud-auth-plugin ==="

        if command -v gke-gcloud-auth-plugin >/dev/null 2>&1; then
            echo "  gke-gcloud-auth-plugin already installed: $(gke-gcloud-auth-plugin --version)"
            exit 0
        fi

        # Try gcloud components install first (works for non-managed installs)
        if gcloud components install gke-gcloud-auth-plugin --quiet 2>/dev/null; then
            echo "  Installed via gcloud components"
        else
            # Add Google Cloud SDK apt repo if not present
            if ! grep -r "packages.cloud.google.com" /etc/apt/sources.list.d/ >/dev/null 2>&1; then
                curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg \
                    | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg
                echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" \
                    | sudo tee /etc/apt/sources.list.d/google-cloud-sdk.list > /dev/null
                sudo apt-get update -y -q
            fi
            sudo apt-get install -y google-cloud-sdk-gke-gcloud-auth-plugin
        fi

        gke-gcloud-auth-plugin --version
        echo "gke-gcloud-auth-plugin installation complete."
    '''
}
return this
