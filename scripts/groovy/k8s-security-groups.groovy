#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_security_group.cluster \
            -target=aws_security_group.nodes \
            -auto-approve -input=false
    """
}

return this
