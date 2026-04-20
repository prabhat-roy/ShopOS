def call() {
    def envVars = [:]
    if (fileExists('infra.env')) {
        readFile('infra.env').trim().split('\n').each { line ->
            def parts = line.split('=', 2)
            if (parts.length == 2) envVars[parts[0]] = parts[1]
        }
    }

    def acrName       = 'shoposregistry'
    def acrSku        = 'Standard'

    // Resolve resource group: infra.env → az account default group → create one
    def resourceGroup = envVars['ACR_RESOURCE_GROUP'] ?: sh(
        script: "az config get defaults.group --query value -o tsv 2>/dev/null || echo ''",
        returnStdout: true
    ).trim()
    if (!resourceGroup) resourceGroup = 'shopos-registry-rg'

    def location = sh(
        script: "az group show --name ${resourceGroup} --query location -o tsv 2>/dev/null || echo ''",
        returnStdout: true
    ).trim()
    if (!location) {
        location = sh(
            script: "az account list-locations --query \"[?isDefault].name\" -o tsv 2>/dev/null | head -1 || echo 'eastus'",
            returnStdout: true
        ).trim() ?: 'eastus'
        sh "az group create --name ${resourceGroup} --location ${location} --output none"
        echo "Created resource group: ${resourceGroup} in ${location}"
    }

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

    sh "sed -i '/^ACR_REGISTRY=/d' infra.env 2>/dev/null || true; echo 'ACR_REGISTRY=${loginServer}' >> infra.env" 
    sh "sed -i '/^ACR_NAME=/d' infra.env 2>/dev/null || true; echo 'ACR_NAME=${acrName}' >> infra.env" 
    sh "sed -i '/^ACR_RESOURCE_GROUP=/d' infra.env 2>/dev/null || true; echo 'ACR_RESOURCE_GROUP=${resourceGroup}' >> infra.env" 
    echo "ACR login server: ${loginServer}"
}

return this
