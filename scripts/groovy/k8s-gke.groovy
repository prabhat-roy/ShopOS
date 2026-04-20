#!/usr/bin/env groovy

def call(String tfDir, String environment = 'dev') {
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

    echo "Provisioning GKE cluster — project=${projectId}  region=${region}  environment=${environment}"

    sh """
        cd ${tfDir}
        terraform apply \
            -var project_id=${projectId} \
            -var region=${region} \
            -var environment=${environment} \
            -auto-approve -input=false
        echo "=== GKE outputs ==="
        terraform output
    """
}

return this
