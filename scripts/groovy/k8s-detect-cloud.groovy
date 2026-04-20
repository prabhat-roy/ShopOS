#!/usr/bin/env groovy

def call() {
    def tfDirMap = [
        AWS   : 'infra/terraform/aws/eks',
        GCP   : 'infra/terraform/gcp/gke',
        AZURE : 'infra/terraform/azure/aks',
    ]

    def detector = load 'scripts/groovy/CloudProviderDetector.groovy'
    detector.detectAndSaveCloudProvider('infra.env')

    def cloud = env.CLOUD_PROVIDER
    def tfDir = tfDirMap[cloud]

    if (!tfDir) {
        error "Unsupported or unknown cloud provider: ${cloud}"
    }

    def lines = fileExists('infra.env') ? readFile('infra.env').readLines() : []
    def hasTfDir = lines.any { it.startsWith('TF_DIR=') }

    if (hasTfDir) {
        lines = lines.collect { it.startsWith('TF_DIR=') ? "TF_DIR=${tfDir}" : it }
    } else {
        lines.add("TF_DIR=${tfDir}")
    }

    writeFile file: 'infra.env', text: lines.join('\n') + '\n'
    echo "TF_DIR=${tfDir}"
}

return this
