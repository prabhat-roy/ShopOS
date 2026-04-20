#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform destroy -auto-approve -input=false
    """
}

return this
