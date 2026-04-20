def call() {
    def envVars = readFile('infra.env').trim().split('\n').collectEntries { line ->
        def parts = line.split('=', 2)
        parts.length == 2 ? [(parts[0]): parts[1]] : [:]
    }

    def tfDir     = envVars['TF_DIR'] ?: 'infra/terraform/gcp/gke'
    def projectId = sh(script: "cd ${tfDir} && terraform output -raw project_id", returnStdout: true).trim()
    def region    = sh(script: "cd ${tfDir} && terraform output -raw region",     returnStdout: true).trim()
    def repoId    = 'shopos'

    sh """
        gcloud artifacts repositories delete ${repoId} \
            --location=${region} \
            --project=${projectId} \
            --quiet
    """
    echo "Deleted Artifact Registry: ${repoId}"
}

return this
