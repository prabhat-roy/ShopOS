pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 90, unit: 'MINUTES')
    }

    parameters {
        choice(
            name: 'ACTION',
            choices: ['BUILD', 'VALIDATE', 'CLEAN'],
            description: 'BUILD — bake new images. VALIDATE — packer validate only. CLEAN — deregister old AMIs/images.'
        )
        choice(
            name: 'TARGET',
            choices: ['jenkins-agent', 'k8s-node-base', 'all'],
            description: 'Which Packer template to bake.'
        )
        choice(
            name: 'CLOUD',
            choices: ['aws', 'gcp', 'both'],
            description: 'Target cloud(s) for the image build.'
        )
        string(
            name: 'GCP_PROJECT_ID',
            defaultValue: 'shopos-prod',
            description: 'GCP project ID (when CLOUD=gcp or both).'
        )
        string(
            name: 'AGE_DAYS',
            defaultValue: '30',
            description: 'When ACTION=CLEAN, deregister images older than N days.'
        )
    }

    environment {
        PACKER_DIR = "infra/packer"
    }

    stages {
        stage('Git Fetch') {
            steps { checkout scm }
        }

        stage('Install Packer') {
            steps {
                sh '''
                    if ! command -v packer >/dev/null 2>&1; then
                        echo "Installing Packer..."
                        wget -qO- https://releases.hashicorp.com/packer/1.11.2/packer_1.11.2_linux_amd64.zip | \
                            unzip -d /tmp/ -
                        sudo install /tmp/packer /usr/local/bin/packer
                    fi
                    packer version
                '''
            }
        }

        stage('Packer Init + Validate') {
            steps {
                script {
                    def targets = params.TARGET == 'all' ? ['jenkins-agent', 'k8s-node-base'] : [params.TARGET]
                    targets.each { tgt ->
                        if (fileExists("${env.PACKER_DIR}/${tgt}/${tgt}.pkr.hcl")) {
                            sh """
                                cd ${env.PACKER_DIR}/${tgt}
                                packer init .
                                packer validate -var gcp_project_id=${params.GCP_PROJECT_ID} ${tgt}.pkr.hcl
                                packer fmt -check . || packer fmt .
                            """
                        } else {
                            echo "Skipping ${tgt} — template not found at ${env.PACKER_DIR}/${tgt}/"
                        }
                    }
                }
            }
        }

        stage('Bake jenkins-agent') {
            when {
                allOf {
                    expression { params.ACTION == 'BUILD' }
                    expression { params.TARGET == 'jenkins-agent' || params.TARGET == 'all' }
                }
            }
            steps {
                script {
                    def only = params.CLOUD == 'aws'  ? '-only=jenkins-agent.amazon-ebs.jenkins-agent' :
                               params.CLOUD == 'gcp'  ? '-only=jenkins-agent.googlecompute.jenkins-agent' : ''
                    sh """
                        cd ${env.PACKER_DIR}/jenkins-agent
                        packer build ${only} \
                            -var gcp_project_id=${params.GCP_PROJECT_ID} \
                            -timestamp-ui \
                            jenkins-agent.pkr.hcl
                    """
                    archiveArtifacts artifacts: "${env.PACKER_DIR}/jenkins-agent/manifest.json", fingerprint: true
                }
            }
        }

        stage('Bake k8s-node-base') {
            when {
                allOf {
                    expression { params.ACTION == 'BUILD' }
                    expression { params.TARGET == 'k8s-node-base' || params.TARGET == 'all' }
                    expression { fileExists("${env.PACKER_DIR}/k8s-node-base/k8s-node-base.pkr.hcl") }
                }
            }
            steps {
                script {
                    def only = params.CLOUD == 'aws'  ? '-only=k8s-node.amazon-ebs.k8s-node' :
                               params.CLOUD == 'gcp'  ? '-only=k8s-node.googlecompute.k8s-node' : ''
                    sh """
                        cd ${env.PACKER_DIR}/k8s-node-base
                        packer build ${only} \
                            -var gcp_project_id=${params.GCP_PROJECT_ID} \
                            -timestamp-ui \
                            k8s-node-base.pkr.hcl
                    """
                    archiveArtifacts artifacts: "${env.PACKER_DIR}/k8s-node-base/manifest.json", fingerprint: true
                }
            }
        }

        stage('Clean Old Images') {
            when { expression { params.ACTION == 'CLEAN' } }
            steps {
                sh """
                    AGE_DAYS=${params.AGE_DAYS}
                    CUTOFF=\$(date -u -d "\${AGE_DAYS} days ago" +%Y-%m-%dT%H:%M:%S)
                    if [ "${params.CLOUD}" = "aws" ] || [ "${params.CLOUD}" = "both" ]; then
                        echo "=== Deregistering AWS AMIs older than \${CUTOFF} ==="
                        aws ec2 describe-images --owners self \
                            --filters "Name=tag:ManagedBy,Values=packer" \
                            --query "Images[?CreationDate<'\${CUTOFF}'].[ImageId,Name,CreationDate]" \
                            --output text | while read ami name date; do
                            [ -z "\$ami" ] && continue
                            echo "Deregistering \$ami (\$name, \$date)"
                            aws ec2 deregister-image --image-id "\$ami" || true
                        done
                    fi
                    if [ "${params.CLOUD}" = "gcp" ] || [ "${params.CLOUD}" = "both" ]; then
                        echo "=== Deleting GCE images older than \${CUTOFF} ==="
                        gcloud compute images list --project=${params.GCP_PROJECT_ID} \
                            --filter="labels.managed=packer AND creationTimestamp<\${CUTOFF}" \
                            --format="value(name)" | while read img; do
                            [ -z "\$img" ] && continue
                            echo "Deleting \$img"
                            gcloud compute images delete "\$img" --project=${params.GCP_PROJECT_ID} --quiet || true
                        done
                    fi
                """
            }
        }

        stage('Publish Image IDs') {
            when { expression { params.ACTION == 'BUILD' } }
            steps {
                sh '''
                    echo "=== Latest Packer artifacts ==="
                    find infra/packer -name manifest.json -exec sh -c \
                        'echo "--- $1 ---"; cat "$1"' _ {} \\;
                    # Persist IDs into infra.env so other pipelines can pin
                    if [ -f infra/packer/jenkins-agent/manifest.json ]; then
                        AMI=$(jq -r '.builds[-1].artifact_id' infra/packer/jenkins-agent/manifest.json | cut -d':' -f2 | head -1)
                        sed -i '/^JENKINS_AGENT_AMI=/d' infra.env 2>/dev/null || true
                        echo "JENKINS_AGENT_AMI=$AMI" >> infra.env
                    fi
                '''
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
        }
        success { echo "Packer ${params.ACTION} for ${params.TARGET} on ${params.CLOUD} completed." }
        failure { echo "Packer ${params.ACTION} failed — check stage logs." }
    }
}
