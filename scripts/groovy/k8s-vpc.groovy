#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_vpc.this \
            -auto-approve -input=false
    """
}

return this
