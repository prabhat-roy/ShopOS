#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v docker >/dev/null 2>&1; then
            echo "Docker already installed: $(docker --version)"
        else
            sudo apt-get update -y
            sudo apt-get install -y ca-certificates gnupg lsb-release

            sudo install -m 0755 -d /etc/apt/keyrings
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
                | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
            sudo chmod a+r /etc/apt/keyrings/docker.gpg

            echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
                | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

            sudo apt-get update -y
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io \
                docker-buildx-plugin docker-compose-plugin

            sudo usermod -aG docker jenkins
            sudo systemctl enable docker
            sudo systemctl start docker

            echo "Docker installed: $(docker --version)"
        fi
    '''
}

return this
