#!/usr/bin/env groovy

def call(String tfDir) {
    def cloud = ''
    if (fileExists('infra.env')) {
        cloud = readFile('infra.env').trim().split('\n').find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: ''
    }

    // AWS: networking stages already ran (VPC/Subnets/IGW/NAT/etc) so target only the cluster.
    // GCP/Azure: no separate networking stages — apply the full plan (VPC + cluster together).
    if (cloud == 'AWS') {
        sh """
            cd ${tfDir}
            terraform apply \
                -target=aws_eks_cluster.this \
                -auto-approve -input=false
            echo "--- EKS outputs ---"
            terraform output
        """
    } else if (cloud == 'GCP') {
        sh """
            cd ${tfDir}
            terraform apply -auto-approve -input=false
            echo "--- GKE outputs ---"
            terraform output
        """
    } else if (cloud == 'AZURE') {
        sh """
            cd ${tfDir}
            terraform apply -auto-approve -input=false
            echo "--- AKS outputs ---"
            terraform output
        """
    } else {
        error "Unsupported cloud provider for cluster apply: '${cloud}'"
    }
}

return this
