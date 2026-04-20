#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_subnet.public \
            -target=aws_subnet.private \
            -auto-approve -input=false
    """
}

return this
