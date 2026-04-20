#!/usr/bin/env groovy

def call(String tfDir) {
    sh """
        cd ${tfDir}
        terraform apply \
            -target=aws_eks_cluster.this \
            -auto-approve -input=false

        echo "--- EKS Cluster outputs ---"
        terraform output
    """
}

return this
