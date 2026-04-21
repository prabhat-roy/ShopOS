#!/usr/bin/env groovy

def call(String tfDir) {
    if (!fileExists('infra.env')) error "infra.env not found — nothing to destroy"

    def props = readFile('infra.env').trim().split('\n').collectEntries { line ->
        def parts = line.split('=', 2)
        parts.length == 2 ? [(parts[0]): parts[1]] : [:]
    }
    def cloud       = props['CLOUD_PROVIDER'] ?: ''
    def environment = props['ENVIRONMENT']    ?: 'dev'
    echo "Destroying ${cloud} cluster (environment=${environment})"

    if (cloud == 'GCP') {
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
        if (!projectId) error "Cannot determine GCP project ID for destroy"

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
            if (zone && zone.contains('-')) region = zone.substring(0, zone.lastIndexOf('-'))
        }
        if (!region) region = 'us-central1'

        sh """
            cd ${tfDir}
            terraform destroy \
                -var project_id=${projectId} \
                -var region=${region} \
                -var environment=${environment} \
                -auto-approve -input=false
        """

    } else if (cloud == 'AZURE') {
        def subscriptionId = sh(
            script: "az account show --query id -o tsv 2>/dev/null || echo ''",
            returnStdout: true
        ).trim()
        if (!subscriptionId) error "Cannot determine Azure subscription ID for destroy"

        sh """
            cd ${tfDir}
            terraform destroy \
                -var subscription_id=${subscriptionId} \
                -var environment=${environment} \
                -auto-approve -input=false
        """

    } else {
        sh """
            cd ${tfDir}
            terraform destroy \
                -var environment=${environment} \
                -auto-approve -input=false
        """
    }
}

return this
