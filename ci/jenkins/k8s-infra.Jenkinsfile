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
                }
            }
        }

        stage('Terraform Init') {
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def tfInit = load 'scripts/groovy/k8s-tf-init.groovy'
                    tfInit(tfDir)
                }
            }
        }

        stage('VPC') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def vpc = load 'scripts/groovy/k8s-vpc.groovy'
                    vpc(tfDir)
                }
            }
        }

        stage('Subnets') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def subnets = load 'scripts/groovy/k8s-subnets.groovy'
                    subnets(tfDir)
                }
            }
        }

        stage('Internet Gateway') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def igw = load 'scripts/groovy/k8s-igw.groovy'
                    igw(tfDir)
                }
            }
        }

        stage('NAT Gateway') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def nat = load 'scripts/groovy/k8s-nat-gateway.groovy'
                    nat(tfDir)
                }
            }
        }

        stage('Route Tables') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def rt = load 'scripts/groovy/k8s-route-tables.groovy'
                    rt(tfDir)
                }
            }
        }

        stage('Security Groups') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def sg = load 'scripts/groovy/k8s-security-groups.groovy'
                    sg(tfDir)
                }
            }
        }

        stage('IAM Roles') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def iam = load 'scripts/groovy/k8s-iam.groovy'
                    iam(tfDir)
                }
            }
        }

        stage('Kubernetes Cluster') {
            when { expression { params.ACTION == 'CREATE' } }
            steps {
                script {
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def cluster = load 'scripts/groovy/k8s-cluster.groovy'
                    cluster(tfDir)
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
                    def tfDir = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('TF_DIR=') }?.split('=', 2)?.last()
                    def destroy = load 'scripts/groovy/k8s-destroy.groovy'
                    destroy(tfDir)
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
