pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 60, unit: 'MINUTES')
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
            }
        }

        stage('Configure Java') {
            steps {
                script {
                    def installJava = load 'scripts/groovy/install-java.groovy'
                    installJava()
                }
            }
        }

        stage('Install curl') {
            steps {
                script {
                    def installCurl = load 'scripts/groovy/install-curl.groovy'
                    installCurl()
                }
            }
        }

        stage('Install zip') {
            steps {
                script {
                    def installZip = load 'scripts/groovy/install-zip.groovy'
                    installZip()
                }
            }
        }

        stage('Install Git') {
            steps {
                script {
                    def installGit = load 'scripts/groovy/install-git.groovy'
                    installGit()
                }
            }
        }

        stage('Install Docker') {
            steps {
                script {
                    def installDocker = load 'scripts/groovy/install-docker.groovy'
                    installDocker()
                }
            }
        }

        stage('Install kubectl') {
            steps {
                script {
                    def installKubectl = load 'scripts/groovy/install-kubectl.groovy'
                    installKubectl()
                }
            }
        }

        stage('Install Helm') {
            steps {
                script {
                    def installHelm = load 'scripts/groovy/install-helm.groovy'
                    installHelm()
                }
            }
        }

        stage('Install Python') {
            steps {
                script {
                    def installPython = load 'scripts/groovy/install-python.groovy'
                    installPython()
                }
            }
        }

        stage('Install Maven') {
            steps {
                script {
                    def installMaven = load 'scripts/groovy/install-maven.groovy'
                    installMaven()
                }
            }
        }

        stage('Install Gradle') {
            steps {
                script {
                    def installGradle = load 'scripts/groovy/install-gradle.groovy'
                    installGradle()
                }
            }
        }

        stage('Install Node.js') {
            steps {
                script {
                    def installNodejs = load 'scripts/groovy/install-nodejs.groovy'
                    installNodejs()
                }
            }
        }

        stage('Install Go') {
            steps {
                script {
                    def installGo = load 'scripts/groovy/install-go.groovy'
                    installGo()
                }
            }
        }

        stage('Install Rust') {
            steps {
                script {
                    def installRust = load 'scripts/groovy/install-rust.groovy'
                    installRust()
                }
            }
        }

        stage('Install .NET') {
            steps {
                script {
                    def installDotnet = load 'scripts/groovy/install-dotnet.groovy'
                    installDotnet()
                }
            }
        }

        stage('Install sbt') {
            steps {
                script {
                    def installSbt = load 'scripts/groovy/install-sbt.groovy'
                    installSbt()
                }
            }
        }

        stage('Install Kotlin') {
            steps {
                script {
                    def installKotlin = load 'scripts/groovy/install-kotlin.groovy'
                    installKotlin()
                }
            }
        }

        stage('Install Terraform') {
            steps {
                script {
                    def installTerraform = load 'scripts/groovy/install-terraform.groovy'
                    installTerraform()
                }
            }
        }

        stage('Install Cosign') {
            steps {
                script {
                    def installCosign = load 'scripts/groovy/install-cosign.groovy'
                    installCosign()
                }
            }
        }

        stage('Install Notation') {
            steps {
                script {
                    def installNotation = load 'scripts/groovy/install-notation.groovy'
                    installNotation()
                }
            }
        }

        stage('Install Skopeo') {
            steps {
                script {
                    def installSkopeo = load 'scripts/groovy/install-skopeo.groovy'
                    installSkopeo()
                }
            }
        }

        stage('Install AWS CLI') {
            steps {
                script {
                    def installAwsCli = load 'scripts/groovy/install-aws-cli.groovy'
                    installAwsCli()
                }
            }
        }

        stage('Install gcloud CLI') {
            steps {
                script {
                    def installGcloud = load 'scripts/groovy/install-gcloud.groovy'
                    installGcloud()
                }
            }
        }

        stage('Install gke-gcloud-auth-plugin') {
            steps {
                script {
                    def installPlugin = load 'scripts/groovy/install-gke-gcloud-auth-plugin.groovy'
                    installPlugin()
                }
            }
        }

        stage('Install Azure CLI') {
            steps {
                script {
                    def installAzCli = load 'scripts/groovy/install-az-cli.groovy'
                    installAzCli()
                }
            }
        }
    }

    post {
        success {
            echo "All tools installed successfully."
        }
        failure {
            echo "Tool installation failed — check stage logs above."
        }
        cleanup {
            sh 'rm -f /tmp/install-*.sh /tmp/install-*.tar.gz /tmp/install-*.zip 2>/dev/null || true'
        }
    }
}
