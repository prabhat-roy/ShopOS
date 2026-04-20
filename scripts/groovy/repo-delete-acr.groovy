def call() {
    def envVars = [:]
    if (fileExists('infra.env')) {
        envVars = readFile('infra.env').trim().split('\n').collectEntries { line ->
            def parts = line.split('=', 2)
            parts.length == 2 ? [(parts[0]): parts[1]] : [:]
        }
    }

    def acrName       = 'shopos'
    def resourceGroup = envVars['ACR_RESOURCE_GROUP'] ?: sh(
        script: "az acr show --name ${acrName} --query resourceGroup -o tsv 2>/dev/null || echo ''",
        returnStdout: true
    ).trim()

    if (resourceGroup) {
        sh "az acr delete --name ${acrName} --resource-group ${resourceGroup} --yes"
        echo "Deleted ACR: ${acrName}"
    } else {
        echo "ACR ${acrName} not found — skipping"
    }
}

return this
