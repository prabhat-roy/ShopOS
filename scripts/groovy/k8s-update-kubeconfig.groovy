#!/usr/bin/env groovy

def call(String envFile = 'infra.env') {
    if (!fileExists(envFile)) error "infra.env not found — run Detect Cloud Provider stage first"
    def props = readFile(envFile).trim().split('\n').collectEntries { line ->
        def parts = line.split('=', 2)
        parts.length == 2 ? [(parts[0]): parts[1]] : [:]
    }

    def cloud  = props['CLOUD_PROVIDER']
    def tfDir  = props['TF_DIR']
    def kubeconfigPath = "${env.WORKSPACE}/kubeconfig"

    def tfOutput = { String key ->
        sh(script: "terraform -chdir=${tfDir} output -raw ${key} 2>/dev/null || echo ''",
           returnStdout: true).trim()
    }

    if (cloud == 'AWS') {
        def clusterName = tfOutput('cluster_name')
        def region      = tfOutput('region') ?: 'us-east-1'
        sh """
            aws eks update-kubeconfig \
                --region ${region} \
                --name ${clusterName} \
                --kubeconfig ${kubeconfigPath}
        """
    } else if (cloud == 'GCP') {
        def clusterName = tfOutput('cluster_name')
        def region      = tfOutput('region')
        def project     = tfOutput('project_id')
        sh """
            KUBECONFIG=${kubeconfigPath} gcloud container clusters get-credentials \
                ${clusterName} \
                --region ${region} \
                --project ${project}
        """
    } else if (cloud == 'AZURE') {
        def clusterName     = tfOutput('cluster_name')
        def resourceGroup   = tfOutput('resource_group_name')
        sh """
            az aks get-credentials \
                --resource-group ${resourceGroup} \
                --name ${clusterName} \
                --file ${kubeconfigPath} \
                --overwrite-existing
        """
    } else {
        error "Unsupported cloud provider for kubeconfig update: ${cloud}"
    }

    sh "kubectl --kubeconfig=${kubeconfigPath} cluster-info"

    def kubeconfigContent = sh(
        script: "base64 -w 0 ${kubeconfigPath}",
        returnStdout: true
    ).trim()

    def entries = [
        KUBECONFIG_PATH   : kubeconfigPath,
        KUBECONFIG_CONTENT: kubeconfigContent,
    ]

    def lines = readFile(envFile).readLines()
    entries.each { key, value ->
        def updated = false
        lines = lines.collect { line ->
            if (line.startsWith("${key}=")) {
                updated = true
                return "${key}=${value}"
            }
            return line
        }
        if (!updated) lines.add("${key}=${value}")
    }

    writeFile file: envFile, text: lines.join('\n') + '\n'
    echo "KUBECONFIG_PATH and KUBECONFIG_CONTENT saved to ${envFile}"
}

return this
