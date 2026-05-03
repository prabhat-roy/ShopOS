#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v git >/dev/null 2>&1; then
            echo "git already installed: $(git --version)"
        else
            sudo apt-get update -y
            sudo apt-get install -y git
            echo "git installed: $(git --version)"
        fi
    '''
}

return this
