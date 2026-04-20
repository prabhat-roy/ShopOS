def call() {
    sh '''
        echo "=== Install AWS CLI ==="

        if command -v aws >/dev/null 2>&1; then
            echo "  AWS CLI already installed: $(aws --version)"
            return 0
        fi

        curl -sf "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o /tmp/awscliv2.zip
        unzip -q /tmp/awscliv2.zip -d /tmp/awscli-install
        /tmp/awscli-install/aws/install --update
        rm -rf /tmp/awscliv2.zip /tmp/awscli-install

        aws --version
        echo "AWS CLI installed."
    '''
}
return this
