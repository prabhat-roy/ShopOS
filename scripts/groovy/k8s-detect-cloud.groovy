#!/usr/bin/env groovy

def call() {
    def iacTool = env.IaC_TOOL ?: 'terraform'
    def tfDirMap = (iacTool == 'opentofu') ? [
        AWS   : 'infra/opentofu/aws/app-k8s',
        GCP   : 'infra/opentofu/gcp/app-k8s',
        AZURE : 'infra/opentofu/azure/app-k8s',
    ] : [
        AWS   : 'infra/terraform/aws/app-k8s',
        GCP   : 'infra/terraform/gcp/app-k8s',
        AZURE : 'infra/terraform/azure/app-k8s',
    ]

    // Re-use previously detected cloud — only detect if not already in infra.env
    def cloud = ''
    if (fileExists('infra.env')) {
        cloud = readFile('infra.env').trim().split('\n')
            .find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: ''
    }

    if (cloud) {
        echo "CLOUD_PROVIDER=${cloud} (loaded from infra.env — skipping detection)"
        env.CLOUD_PROVIDER = cloud
    } else {
        def detector = load 'scripts/groovy/CloudProviderDetector.groovy'
        detector.detectAndSaveCloudProvider('infra.env')
        cloud = env.CLOUD_PROVIDER
    }
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
