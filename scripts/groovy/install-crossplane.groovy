#!/usr/bin/env groovy

def call() {
    sh """
        # Install Crossplane via Helm
        helm repo add crossplane-stable https://charts.crossplane.io/stable || true
        helm repo update
        helm upgrade --install crossplane crossplane-stable/crossplane \
            --namespace crossplane-system \
            --create-namespace \
            --set args='{--debug}' \
            --set resourcesCrossplane.limits.cpu=500m \
            --set resourcesCrossplane.limits.memory=512Mi \
            --set resourcesCrossplane.requests.cpu=100m \
            --set resourcesCrossplane.requests.memory=256Mi \
            --wait --timeout 5m
    """
    sh """
        # Wait for Crossplane to be ready
        kubectl rollout status deployment/crossplane -n crossplane-system --timeout=3m

        # Apply compositions and claims from infra/crossplane/
        kubectl apply -f infra/crossplane/crossplane-install.yaml --server-side || true
        kubectl apply -f infra/crossplane/compositions/ || true
        echo "Crossplane installed — compositions applied"
    """
    echo 'Crossplane installed'
}

return this
