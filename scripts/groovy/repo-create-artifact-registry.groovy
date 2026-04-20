def call() {
    def envVars = readFile('infra.env').trim().split('\n').collectEntries { line ->
        def parts = line.split('=', 2)
        parts.length == 2 ? [(parts[0]): parts[1]] : [:]
    }

    def tfDir = envVars['TF_DIR'] ?: 'infra/terraform/gcp/gke'
    def projectId = sh(script: "cd ${tfDir} && terraform output -raw project_id", returnStdout: true).trim()
    def region    = sh(script: "cd ${tfDir} && terraform output -raw region",     returnStdout: true).trim()

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
