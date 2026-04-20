#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v python3 >/dev/null 2>&1 && command -v pip3 >/dev/null 2>&1; then
            echo "Python already installed: $(python3 --version)"
            echo "pip: $(pip3 --version)"
        else
            sudo apt-get update -y
            sudo apt-get install -y python3 python3-pip python3-venv python3-dev
            echo "Python installed: $(python3 --version)"
            echo "pip: $(pip3 --version)"
        fi
    '''
}

return this
