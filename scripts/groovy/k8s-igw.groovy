#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_internet_gateway.this \
            -auto-approve -input=false
    """
}

return this
