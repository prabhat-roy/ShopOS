def call() {
    // Resolve project_id: gcloud config → GCE metadata → error
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

    // Resolve region: gcloud config → GCE metadata zone stripped to region → default
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
        // Strip trailing zone letter: us-central1-c → us-central1
        if (zone && zone.contains('-')) {
            region = zone.substring(0, zone.lastIndexOf('-'))
        }
    }
    if (!region) region = 'us-central1'

    def repoId = 'shopos'

    def existing = sh(
        script: "gcloud artifacts repositories describe ${repoId} --location=${region} --project=${projectId} 2>&1 || true",
        returnStdout: true
    ).trim()

    if (!existing.contains('name:')) {
        sh """
            gcloud artifacts repositories create ${repoId} \
                --repository-format=docker \
                --location=${region} \
                --project=${projectId} \
                --description='ShopOS container images'
        """
        echo "Created Artifact Registry: ${repoId} in ${region}"
    } else {
        echo "Artifact Registry already exists: ${repoId}"
    }

    def registryUrl = "${region}-docker.pkg.dev/${projectId}/${repoId}"
    sh "echo 'ARTIFACT_REGISTRY=${registryUrl}' >> infra.env"
    sh "echo 'ARTIFACT_REGISTRY_REGION=${region}' >> infra.env"
    sh "echo 'ARTIFACT_REGISTRY_PROJECT=${projectId}' >> infra.env"
    echo "Artifact Registry: ${registryUrl}"
}

return this
