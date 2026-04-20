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
                    // Cache TF_DIR in env so all subsequent stages reuse without re-reading the file
                    env.TF_DIR = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    echo "TF_DIR=${env.TF_DIR}"
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

        stage('VPC') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def vpc = load 'scripts/groovy/k8s-vpc.groovy'
                    vpc(env.TF_DIR)
                }
            }
        }

        stage('Subnets') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def subnets = load 'scripts/groovy/k8s-subnets.groovy'
                    subnets(env.TF_DIR)
                }
            }
        }

        stage('Internet Gateway') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def igw = load 'scripts/groovy/k8s-igw.groovy'
                    igw(env.TF_DIR)
                }
            }
        }

        stage('NAT Gateway') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def nat = load 'scripts/groovy/k8s-nat-gateway.groovy'
                    nat(env.TF_DIR)
                }
            }
        }

        stage('Route Tables') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def rt = load 'scripts/groovy/k8s-route-tables.groovy'
                    rt(env.TF_DIR)
                }
            }
        }

        stage('Security Groups') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def sg = load 'scripts/groovy/k8s-security-groups.groovy'
                    sg(env.TF_DIR)
                }
            }
        }

        stage('IAM Roles') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def iam = load 'scripts/groovy/k8s-iam.groovy'
                    iam(env.TF_DIR)
                }
            }
        }

        stage('Kubernetes Cluster') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def cluster = load 'scripts/groovy/k8s-cluster.groovy'
                    cluster(env.TF_DIR)
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
