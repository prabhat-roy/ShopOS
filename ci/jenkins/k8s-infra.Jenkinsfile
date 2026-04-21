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
            choices: ['CREATE', 'DESTROY'],
            description: 'Create or destroy the Kubernetes cluster infrastructure'
        )
        choice(
            name: 'ENVIRONMENT',
            choices: ['dev', 'staging', 'prod'],
            description: 'Target environment — passed to Terraform as var.environment'
        )
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
            }
        }

        stage('Detect Cloud Provider') {
            steps {
                script {
                    def detectCloud = load 'scripts/groovy/k8s-detect-cloud.groovy'
                    detectCloud()
                    env.TF_DIR = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    env.CLOUD_PROVIDER = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last()
                    echo "CLOUD_PROVIDER=${env.CLOUD_PROVIDER}  TF_DIR=${env.TF_DIR}"
                }
            }
        }

        stage('Terraform Init') {
            steps {
                script {
                    def tfInit = load 'scripts/groovy/k8s-tf-init.groovy'
                    tfInit(env.TF_DIR)
                }
            }
        }

        // ── CREATE path ───────────────────────────────────────────────────────
        stage('Create Infrastructure') {
            when {
                beforeAgent true
                expression { params.ACTION == 'CREATE' }
            }
            stages {
                stage('Provision Cluster') {
                    steps {
                        script {
                            if (env.CLOUD_PROVIDER == 'GCP') {
                                def s = load 'scripts/groovy/k8s-gke.groovy'
                                s(env.TF_DIR, params.ENVIRONMENT)
                            } else if (env.CLOUD_PROVIDER == 'AWS') {
                                def s = load 'scripts/groovy/k8s-eks.groovy'
                                s(env.TF_DIR, params.ENVIRONMENT)
                            } else if (env.CLOUD_PROVIDER == 'AZURE') {
                                def s = load 'scripts/groovy/k8s-aks.groovy'
                                s(env.TF_DIR, params.ENVIRONMENT)
                            } else {
                                error "Unknown CLOUD_PROVIDER: ${env.CLOUD_PROVIDER}"
                            }
                        }
                    }
                }

                stage('Update Kubeconfig') {
                    steps {
                        script {
                            def s = load 'scripts/groovy/k8s-update-kubeconfig.groovy'
                            s()
                        }
                    }
                }
            }
        }

        // ── DESTROY path ──────────────────────────────────────────────────────
        stage('Destroy Infrastructure') {
            when {
                beforeAgent true
                expression { params.ACTION == 'DESTROY' }
            }
            steps {
                script {
                    def s = load 'scripts/groovy/k8s-destroy.groovy'
                    s(env.TF_DIR, params.ENVIRONMENT)
                }
            }
        }
    }

    post {
        success {
            echo "${params.ACTION} of Kubernetes infrastructure completed successfully."
        }
        failure {
            echo "${params.ACTION} failed — check stage logs above."
        }
        cleanup {
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
        }
    }
}
