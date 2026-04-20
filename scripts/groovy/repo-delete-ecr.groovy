def call() {
    def region = sh(script: """
        TOKEN=\$(curl -s -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
        curl -s -H "X-aws-ec2-metadata-token: \$TOKEN" http://169.254.169.254/latest/meta-data/placement/region
    """, returnStdout: true).trim()

    def repos = sh(
        script: "aws ecr describe-repositories --region ${region} --query 'repositories[?starts_with(repositoryName, `shopos/`)].repositoryName' --output text",
        returnStdout: true
    ).trim().split('\\s+')

    repos.each { repo ->
        if (repo?.trim()) {
            sh "aws ecr delete-repository --repository-name ${repo} --region ${region} --force"
            echo "Deleted ECR repo: ${repo}"
        }
    }
    echo "ECR cleanup complete for region ${region}"
}

return this
