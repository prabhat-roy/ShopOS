#!/usr/bin/env groovy

def call(String tfDir, String environment = 'dev') {
    def region = sh(
        script: """
            AWS_REGION=\$(aws configure get region 2>/dev/null || echo '')
            if [ -z "\$AWS_REGION" ]; then
                TOKEN=\$(curl -s -X PUT http://169.254.169.254/latest/api/token \
                    -H 'X-aws-ec2-metadata-token-ttl-seconds: 21600')
                AWS_REGION=\$(curl -s -H "X-aws-ec2-metadata-token: \$TOKEN" \
                    http://169.254.169.254/latest/meta-data/placement/region)
            fi
            echo "\$AWS_REGION"
        """,
        returnStdout: true
    ).trim()
    if (!region) region = 'us-east-1'

    echo "Provisioning EKS cluster — region=${region}  environment=${environment}"

    sh """
        cd ${tfDir}
        terraform apply \
            -var region=${region} \
            -var environment=${environment} \
            -auto-approve -input=false
        echo "=== EKS outputs ==="
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
