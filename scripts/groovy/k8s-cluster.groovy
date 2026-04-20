#!/usr/bin/env groovy

def call(String tfDir, String environment = 'dev') {
    def cloud = ''
    if (fileExists('infra.env')) {
        cloud = readFile('infra.env').trim().split('\n').find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: ''
    }

    if (cloud == 'AWS') {
        // Networking stages already applied VPC/Subnets/IGW/NAT/SG/IAM — target only the cluster.
        // All AWS vars have defaults so no -var flags needed.
        sh """
            cd ${tfDir}
            terraform apply \
                -target=aws_eks_cluster.this \
                -var environment=${environment} \
                -auto-approve -input=false
            echo "--- EKS outputs ---"
            terraform output
        """

    } else if (cloud == 'GCP') {
        // Detect project_id — required variable with no default
        def projectId = sh(
            script: "gcloud config get-value project 2>/dev/null || echo ''",
            returnStdout: true
        ).trim()
        if (!projectId) {
            projectId = sh(
                script: """curl -sf -H 'Metadata-Flavor: Google' \
                    http://metadata.google.internal/computeMetadata/v1/project/project-id 2>/dev/null || echo ''""",
                returnStdout: true
            ).trim()
        }
        if (!projectId) error "Cannot determine GCP project ID — set gcloud config or run on GCE"

        // Detect region — use GCE metadata zone → strip zone suffix → default
        def region = sh(
            script: "gcloud config get-value compute/region 2>/dev/null || echo ''",
            returnStdout: true
        ).trim()
        if (!region) {
            def zone = sh(
                script: """curl -sf -H 'Metadata-Flavor: Google' \
                    'http://metadata.google.internal/computeMetadata/v1/instance/zone' 2>/dev/null | \
                    awk -F/ '{print \$NF}'""",
                returnStdout: true
            ).trim()
            if (zone && zone.contains('-')) {
                region = zone.substring(0, zone.lastIndexOf('-'))
            }
        }
        if (!region) region = 'us-central1'

        sh """
            cd ${tfDir}
            terraform apply \
                -var project_id=${projectId} \
                -var region=${region} \
                -var environment=${environment} \
                -auto-approve -input=false
            echo "--- GKE outputs ---"
            terraform output
        """

    } else if (cloud == 'AZURE') {
        // Detect subscription_id — required variable with no default
        def subscriptionId = sh(
            script: "az account show --query id -o tsv 2>/dev/null || echo ''",
            returnStdout: true
        ).trim()
        if (!subscriptionId) error "Cannot determine Azure subscription ID — ensure az login has been run"

        sh """
            cd ${tfDir}
            terraform apply \
                -var subscription_id=${subscriptionId} \
                -var environment=${environment} \
                -auto-approve -input=false
            echo "--- AKS outputs ---"
            terraform output
        """

    } else {
        error "Unsupported cloud provider for cluster apply: '${cloud}'"
    }
}

return this
