def call() {
    def envMap = [:]
    if (fileExists('infra.env')) {
        readFile('infra.env').trim().split('\n').each { line ->
            def parts = line.split('=', 2)
            if (parts.length == 2) envMap[parts[0]] = parts[1]
        }
    }

    // Prefer values saved by create step, then fall back to gcloud config
    def projectId = envMap['ARTIFACT_REGISTRY_PROJECT'] ?: sh(
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
    if (!projectId) error "Cannot determine GCP project ID"

    def region = envMap['ARTIFACT_REGISTRY_REGION'] ?: sh(
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

    def repoId = 'shopos'
    sh """
        gcloud artifacts repositories delete ${repoId} \
            --location=${region} \
            --project=${projectId} \
            --quiet
    """
    echo "Deleted Artifact Registry: ${repoId}"
}

return this
