pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 60, unit: 'MINUTES')
    }

    parameters {
        choice(
            name: 'PLAYBOOK',
            choices: ['jenkins', 'k8s-node', 'common'],
            description: 'Ansible playbook to run'
        )
        choice(
            name: 'INVENTORY',
            choices: ['aws', 'gcp', 'azure'],
            description: 'Target cloud inventory'
        )
        string(
            name: 'LIMIT',
            defaultValue: '',
            description: 'Limit to specific host or group (optional, e.g. role_jenkins)'
        )
        string(
            name: 'EXTRA_VARS',
            defaultValue: '',
            description: 'Extra variables to pass to ansible-playbook (e.g. jenkins_admin_password=secret)'
        )
        booleanParam(
            name: 'CHECK_MODE',
            defaultValue: false,
            description: 'Run in check mode (dry run) — no changes will be made'
        )
    }

    environment {
        ANSIBLE_DIR        = 'infra/ansible'
        ANSIBLE_CONFIG     = 'infra/ansible/ansible.cfg'
        ANSIBLE_FORCE_COLOR = '1'
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
            }
        }

        stage('Load Infra Env') {
            steps {
                script {
                    if (fileExists('infra.env')) {
                        def props = readFile('infra.env').trim()
                        props.split('\n').each { line ->
                            if (line && !line.startsWith('#')) {
                                def parts = line.split('=', 2)
                                if (parts.size() == 2) {
                                    env[parts[0].trim()] = parts[1].trim()
                                }
                            }
                        }
                        echo "Loaded infra.env"
                    } else {
                        echo "No infra.env found — using static inventory"
                    }
                }
            }
        }

        stage('Install Ansible') {
            steps {
                sh '''
                    if ! command -v ansible &>/dev/null; then
                        apt-get update -y
                        apt-get install -y ansible python3-pip
                    fi
                    ansible --version
                '''
            }
        }

        stage('Install Collections') {
            steps {
                sh "ansible-galaxy collection install -r ${ANSIBLE_DIR}/requirements.yml --force"
            }
        }

        stage('Install Cloud SDKs') {
            steps {
                sh '''
                    pip3 install --quiet boto3 google-auth azure-mgmt-compute azure-mgmt-network 2>/dev/null || true
                '''
            }
        }

        stage('Validate Inventory') {
            steps {
                sh "ansible-inventory -i ${ANSIBLE_DIR}/inventory/${params.INVENTORY}.yml --list > /dev/null"
                echo "Inventory ${params.INVENTORY} is valid"
            }
        }

        stage('Run Playbook') {
            steps {
                script {
                    def cmd = [
                        "ansible-playbook",
                        "${ANSIBLE_DIR}/playbooks/${params.PLAYBOOK}.yml",
                        "-i ${ANSIBLE_DIR}/inventory/${params.INVENTORY}.yml",
                        params.LIMIT      ? "--limit '${params.LIMIT}'" : '',
                        params.EXTRA_VARS ? "--extra-vars '${params.EXTRA_VARS}'" : '',
                        params.CHECK_MODE ? '--check --diff' : '',
                        '-v'
                    ].findAll { it }.join(' ')

                    echo "Running: ${cmd}"
                    sh cmd
                }
            }
        }
    }

    post {
        success {
            echo "Ansible playbook '${params.PLAYBOOK}' completed successfully on ${params.INVENTORY}"
        }
        failure {
            echo "Ansible playbook '${params.PLAYBOOK}' failed on ${params.INVENTORY}"
        }
    }
}
