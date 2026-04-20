#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v helm >/dev/null 2>&1; then
            echo "Helm already installed: $(helm version --short)"
        else
            curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
            echo "Helm installed: $(helm version --short)"
        fi
    '''
}

return this
