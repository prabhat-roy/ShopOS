#!/usr/bin/env groovy

def call(String tfDir, String environment = 'dev') {
    def subscriptionId = sh(
        script: "az account show --query id -o tsv 2>/dev/null || echo ''",
        returnStdout: true
    ).trim()
    if (!subscriptionId) error "Cannot determine Azure subscription ID — ensure az login has been run"

    def location = sh(
        script: "az account show --query 'environmentName' -o tsv 2>/dev/null || echo 'East US'",
        returnStdout: true
    ).trim() ?: 'East US'

    echo "Provisioning AKS cluster — subscription=${subscriptionId}  environment=${environment}"

    sh """
        cd ${tfDir}
        terraform apply \
            -var subscription_id=${subscriptionId} \
            -var environment=${environment} \
            -auto-approve -input=false
        echo "=== AKS outputs ==="
        terraform output
    """

    def lines = fileExists('infra.env') ? readFile('infra.env').readLines() : []
    def updated = false
    lines = lines.collect { line ->
        if (line.startsWith('ENVIRONMENT=')) { updated = true; return "ENVIRONMENT=${environment}" }
        return line
    }
    if (!updated) lines.add("ENVIRONMENT=${environment}")
    writeFile file: 'infra.env', text: lines.join('\n') + '\n'
}

return this
