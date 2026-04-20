class CloudProviderDetector implements Serializable {
    def steps

    CloudProviderDetector(steps) {
        this.steps = steps
    }

    String detectCloudProvider() {
        try {
            steps.echo "Detecting Cloud Provider..."

            // GCP — distinct hostname; check first to avoid 169.254.169.254 collision
            try {
                def code = steps.sh(
                    script: 'curl -s --max-time 3 -o /dev/null -w "%{http_code}" ' +
                            '-H "Metadata-Flavor: Google" ' +
                            'http://metadata.google.internal/computeMetadata/v1/instance/id',
                    returnStdout: true
                ).trim()
                if (code == '200') {
                    steps.echo "Detected GCP"
                    return 'GCP'
                }
            } catch (Exception ignored) {}

            // AWS — IMDSv2: PUT to get a session token, then GET metadata
            try {
                def token = steps.sh(
                    script: 'curl -s --max-time 3 -X PUT ' +
                            '"http://169.254.169.254/latest/api/token" ' +
                            '-H "X-aws-ec2-metadata-token-ttl-seconds: 10"',
                    returnStdout: true
                ).trim()
                if (token) {
                    def code = steps.sh(
                        script: "curl -s --max-time 3 -o /dev/null -w \"%{http_code}\" " +
                                "-H \"X-aws-ec2-metadata-token: ${token}\" " +
                                "http://169.254.169.254/latest/meta-data/",
                        returnStdout: true
                    ).trim()
                    if (code == '200') {
                        steps.echo "Detected AWS"
                        return 'AWS'
                    }
                }
            } catch (Exception ignored) {}

            // Azure — IMDS requires Metadata header
            try {
                def code = steps.sh(
                    script: 'curl -s --max-time 3 -o /dev/null -w "%{http_code}" ' +
                            '-H "Metadata: true" ' +
                            '"http://169.254.169.254/metadata/instance?api-version=2021-02-01"',
                    returnStdout: true
                ).trim()
                if (code == '200') {
                    steps.echo "Detected Azure"
                    return 'AZURE'
                }
            } catch (Exception ignored) {}

            steps.echo "Cloud Provider Unknown — no IMDS endpoint responded"
            return 'UNKNOWN'

        } catch (Exception e) {
            steps.echo "Error detecting cloud provider: ${e.message}"
            return 'UNKNOWN'
        }
    }

    void detectAndSaveCloudProvider(String envFile = 'infra.env') {
        def cloud = detectCloudProvider()
        steps.echo "Cloud Provider: ${cloud}"

        steps.env.CLOUD_PROVIDER = cloud

        def lines = []
        def updatedCloud = false

        if (steps.fileExists(envFile)) {
            lines = steps.readFile(envFile).readLines()
            lines = lines.collect { line ->
                if (line.startsWith('CLOUD_PROVIDER=')) {
                    updatedCloud = true
                    return "CLOUD_PROVIDER=${cloud}"
                }
                return line
            }
        }

        if (!updatedCloud) {
            lines.add("CLOUD_PROVIDER=${cloud}")
        }

        steps.writeFile file: envFile, text: lines.join('\n') + '\n'
        steps.echo "CLOUD_PROVIDER=${cloud} saved to ${envFile}"
    }
}

return new CloudProviderDetector(this)
