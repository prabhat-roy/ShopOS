#!/usr/bin/env groovy

def call() {
    sh '''
        if command -v terraform >/dev/null 2>&1; then
            echo "Terraform already installed: $(terraform version | head -1)"
        else
            TERRAFORM_VERSION=$(curl -s https://checkpoint-api.hashicorp.com/v1/check/terraform \
                | python3 -c 'import sys,json; print(json.load(sys.stdin)["current_version"])')
            curl -sLO "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
            sudo unzip -q "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" -d /usr/local/bin/
            sudo chmod +x /usr/local/bin/terraform
            rm -f "terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
            echo "Terraform installed: $(terraform version | head -1)"
        fi
    '''

    sh '''
        sudo tee /etc/profile.d/terraform.sh > /dev/null << EOF
export PATH=\\$PATH:/usr/local/bin
EOF
        sudo chmod 644 /etc/profile.d/terraform.sh
        echo "TERRAFORM_BIN=$(which terraform)"
        terraform version | head -1
    '''
}

return this
