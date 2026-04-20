#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform init -input=false
        terraform validate
    """
}

return this
