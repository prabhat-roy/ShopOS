#!/usr/bin/env groovy

def call(String tfDir) {
    def iacCmd = env.IaC_TOOL == 'opentofu' ? 'tofu' : 'terraform'
    def cloud = ''
    if (fileExists('infra.env')) {
        cloud = readFile('infra.env').trim().split('\n')
            .find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: ''
    }

    if (cloud == 'GCP') {
        // Auto-create GCS bucket for remote state if it doesn't exist
        def projectId = sh(
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
        if (!projectId) error "Cannot determine GCP project ID for remote backend"

        def region = sh(
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
            if (zone && zone.contains('-')) region = zone.substring(0, zone.lastIndexOf('-'))
        }
        if (!region) region = 'us-central1'

        def bucket = "shopos-tfstate-${projectId}"
        sh """
            gsutil ls -b gs://${bucket} 2>/dev/null || \
            gsutil mb -p ${projectId} -l ${region} -b on gs://${bucket}
            gsutil versioning set on gs://${bucket} 2>/dev/null || true
        """
        sh """
            cd ${tfDir}
            ${iacCmd} init -input=false \
                -backend-config="bucket=${bucket}" \
                -backend-config="prefix=gke" \
                -reconfigure
            ${iacCmd} validate
        """

    } else if (cloud == 'AWS') {
        def region = sh(
            script: """AWS_REGION=\$(aws configure get region 2>/dev/null || echo '')
                [ -z "\$AWS_REGION" ] && { TOKEN=\$(curl -s -X PUT http://169.254.169.254/latest/api/token -H 'X-aws-ec2-metadata-token-ttl-seconds: 21600'); AWS_REGION=\$(curl -s -H "X-aws-ec2-metadata-token: \$TOKEN" http://169.254.169.254/latest/meta-data/placement/region); }
                echo "\$AWS_REGION" """,
            returnStdout: true
        ).trim()
        if (!region) region = 'us-east-1'

        def accountId = sh(
            script: "aws sts get-caller-identity --query Account --output text",
            returnStdout: true
        ).trim()

        def bucket = "shopos-tfstate-${accountId}"
        def table  = "shopos-tfstate-lock"
        sh """
            aws s3api head-bucket --bucket ${bucket} 2>/dev/null || \
            aws s3api create-bucket --bucket ${bucket} --region ${region} \
                ${region == 'us-east-1' ? '' : "--create-bucket-configuration LocationConstraint=${region}"}
            aws s3api put-bucket-versioning --bucket ${bucket} \
                --versioning-configuration Status=Enabled 2>/dev/null || true
            aws s3api put-bucket-encryption --bucket ${bucket} \
                --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}' 2>/dev/null || true
            aws dynamodb create-table --table-name ${table} \
                --attribute-definitions AttributeName=LockID,AttributeType=S \
                --key-schema AttributeName=LockID,KeyType=HASH \
                --billing-mode PAY_PER_REQUEST \
                --region ${region} 2>/dev/null || true
        """
        sh """
            cd ${tfDir}
            ${iacCmd} init -input=false \
                -backend-config="bucket=${bucket}" \
                -backend-config="key=eks/terraform.tfstate" \
                -backend-config="region=${region}" \
                -backend-config="dynamodb_table=${table}" \
                -backend-config="encrypt=true" \
                -reconfigure
            ${iacCmd} validate
        """

    } else if (cloud == 'AZURE') {
        def subscriptionId = sh(
            script: "az account show --query id -o tsv 2>/dev/null || echo ''",
            returnStdout: true
        ).trim()
        if (!subscriptionId) error "Cannot determine Azure subscription ID for remote backend"

        def rgName      = 'shopos-tfstate-rg'
        def accountName = 'shoposterraformstate'
        def containerName = 'tfstate'
        sh """
            az group create --name ${rgName} --location eastus 2>/dev/null || true
            az storage account create \
                --name ${accountName} \
                --resource-group ${rgName} \
                --location eastus \
                --sku Standard_LRS \
                --encryption-services blob 2>/dev/null || true
            az storage container create \
                --name ${containerName} \
                --account-name ${accountName} 2>/dev/null || true
        """
        def accessKey = sh(
            script: "az storage account keys list --account-name ${accountName} --resource-group ${rgName} --query '[0].value' -o tsv",
            returnStdout: true
        ).trim()
        sh """
            cd ${tfDir}
            ${iacCmd} init -input=false \
                -backend-config="resource_group_name=${rgName}" \
                -backend-config="storage_account_name=${accountName}" \
                -backend-config="container_name=${containerName}" \
                -backend-config="key=aks/terraform.tfstate" \
                -backend-config="access_key=${accessKey}" \
                -reconfigure
            ${iacCmd} validate
        """

    } else {
        sh """
            cd ${tfDir}
            ${iacCmd} init -input=false
            ${iacCmd} validate
        """
    }
}

return this
