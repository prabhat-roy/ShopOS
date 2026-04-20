#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v kubectl >/dev/null 2>&1; then
            echo "kubectl already installed: $(kubectl version --client)"
        else
            KUBECTL_VERSION=$(curl -sL https://dl.k8s.io/release/stable.txt)
            curl -sLO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
            sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
            rm -f kubectl
            echo "kubectl installed: $(kubectl version --client)"
        fi
    '''
}

return this
