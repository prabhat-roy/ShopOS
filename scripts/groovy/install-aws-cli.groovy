def call() {
    sh '''
        echo "=== Install AWS CLI ==="

        if command -v aws >/dev/null 2>&1; then
            echo "  AWS CLI already installed: $(aws --version)"
            return 0
        fi

        sudo rm -rf /tmp/awscliv2.zip /tmp/awscli-install
        curl -sf "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o /tmp/awscliv2.zip
        sudo unzip -oq /tmp/awscliv2.zip -d /tmp/awscli-install
        sudo /tmp/awscli-install/aws/install --update
        sudo rm -rf /tmp/awscliv2.zip /tmp/awscli-install

        aws --version
        echo "AWS CLI installed."
    '''
}
return this
