def call() {
    def envVars = readFile('infra.env').trim().split('\n').collectEntries { line ->
        def parts = line.split('=', 2)
        parts.length == 2 ? [(parts[0]): parts[1]] : [:]
    }

    def tfDir = envVars['TF_DIR'] ?: 'infra/terraform/azure/aks'
    def resourceGroup = sh(script: "cd ${tfDir} && terraform output -raw resource_group_name", returnStdout: true).trim()

    def acrName    = 'shopos'
    def acrSku     = 'Standard'
    def location   = sh(
        script: "az group show --name ${resourceGroup} --query location -o tsv",
        returnStdout: true
    ).trim()

    def existing = sh(
        script: "az acr show --name ${acrName} --resource-group ${resourceGroup} 2>&1 || true",
        returnStdout: true
    ).trim()

    if (!existing.contains('"name"')) {
        sh """
            az acr create \
                --name ${acrName} \
                --resource-group ${resourceGroup} \
                --sku ${acrSku} \
                --location ${location} \
                --admin-enabled false
        """
        echo "Created ACR: ${acrName}"
    } else {
        echo "ACR already exists: ${acrName}"
    }

    def loginServer = sh(
        script: "az acr show --name ${acrName} --query loginServer -o tsv",
        returnStdout: true
    ).trim()

    sh "echo 'ACR_REGISTRY=${loginServer}' >> infra.env"
    sh "echo 'ACR_NAME=${acrName}' >> infra.env"
    sh "echo 'ACR_RESOURCE_GROUP=${resourceGroup}' >> infra.env"
    echo "ACR login server: ${loginServer}"
}

return this
