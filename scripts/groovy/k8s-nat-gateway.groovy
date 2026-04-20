#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_eip.nat \
            -target=aws_nat_gateway.this \
            -auto-approve -input=false
    """
}

return this
