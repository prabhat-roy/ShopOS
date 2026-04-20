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

        stage('Provision Cluster') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    if (env.CLOUD_PROVIDER == 'GCP') {
                        def gke = load 'scripts/groovy/k8s-gke.groovy'
                        gke(env.TF_DIR, params.ENVIRONMENT)
                    } else if (env.CLOUD_PROVIDER == 'AWS') {
                        def eks = load 'scripts/groovy/k8s-eks.groovy'
                        eks(env.TF_DIR, params.ENVIRONMENT)
                    } else if (env.CLOUD_PROVIDER == 'AZURE') {
                        def aks = load 'scripts/groovy/k8s-aks.groovy'
                        aks(env.TF_DIR, params.ENVIRONMENT)
                    } else {
                        error "Unknown CLOUD_PROVIDER: ${env.CLOUD_PROVIDER}"
                    }
                }
            }
        }

        stage('Update Kubeconfig') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def updateKubeconfig = load 'scripts/groovy/k8s-update-kubeconfig.groovy'
                    updateKubeconfig()
                }
            }
        }

        stage('Destroy Infrastructure') {
            when { expression { params.ACTION == 'DESTROY' } }
            steps {
                script {
                    def destroy = load 'scripts/groovy/k8s-destroy.groovy'
                    destroy(env.TF_DIR)
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
