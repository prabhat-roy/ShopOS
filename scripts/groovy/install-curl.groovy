#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v curl >/dev/null 2>&1; then
            echo "curl already installed: $(curl --version | head -1)"
        else
            sudo apt-get update -y
            sudo apt-get install -y curl
            echo "curl installed: $(curl --version | head -1)"
        fi
    '''
}

return this
