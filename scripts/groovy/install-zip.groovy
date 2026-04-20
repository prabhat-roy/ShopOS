#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v zip >/dev/null 2>&1; then
            echo "zip already installed: $(zip --version | head -2)"
        else
            sudo apt-get install -y zip unzip
            echo "zip installed: $(zip --version | head -2)"
        fi
    '''
}

return this
