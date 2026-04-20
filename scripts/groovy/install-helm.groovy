#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v helm >/dev/null 2>&1; then
            echo "Helm already installed: $(helm version --short)"
        else
            curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
                -o /tmp/get-helm-3.sh
            sudo bash /tmp/get-helm-3.sh
            rm -f /tmp/get-helm-3.sh
            echo "Helm installed: $(helm version --short)"
        fi
    '''
}

return this
