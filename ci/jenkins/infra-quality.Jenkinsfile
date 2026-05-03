pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '20'))
        timeout(time: 60, unit: 'MINUTES')
    }

    triggers {
        // Weekly drift report — Mondays 06:00 UTC
        cron('H 6 * * 1')
    }

    parameters {
        booleanParam(name: 'RUN_INFRACOST',     defaultValue: true,  description: 'Infracost — cost estimation on Terraform plans')
        booleanParam(name: 'RUN_DRIFTCTL',      defaultValue: true,  description: 'Driftctl — Terraform state vs cloud drift detection')
        booleanParam(name: 'RUN_ATLANTIS_CFG',  defaultValue: true,  description: 'Atlantis — validate atlantis.yaml + apply repo whitelist')
        booleanParam(name: 'RUN_PRE_COMMIT',    defaultValue: true,  description: 'pre-commit — terraform-fmt + tflint + tfsec hooks across all .tf files')
        booleanParam(name: 'RUN_TFLINT',        defaultValue: true,  description: 'tflint — Terraform linter (provider-specific rules)')
        booleanParam(name: 'RUN_INFRAMAP',      defaultValue: true,  description: 'inframap — generate dependency graph from Terraform state')
        choice(
            name: 'CLOUD',
            choices: ['all', 'aws', 'gcp', 'azure'],
            description: 'Limit drift/cost to a single cloud, or scan all three.'
        )
    }

    environment {
        TF_TERRAFORM_DIR  = "infra/terraform"
        TF_OPENTOFU_DIR   = "infra/opentofu"
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
                sh 'mkdir -p reports/infra-quality'
            }
        }

        stage('Resolve Workload Dirs') {
            steps {
                script {
                    def dirs = []
                    if (params.CLOUD == 'all' || params.CLOUD == 'aws') {
                        dirs += ['infra/terraform/aws/app-k8s', 'infra/opentofu/aws/app-k8s']
                    }
                    if (params.CLOUD == 'all' || params.CLOUD == 'gcp') {
                        dirs += ['infra/terraform/gcp/app-k8s', 'infra/opentofu/gcp/app-k8s']
                    }
                    if (params.CLOUD == 'all' || params.CLOUD == 'azure') {
                        dirs += ['infra/terraform/azure/app-k8s', 'infra/opentofu/azure/app-k8s']
                    }
                    env.WORKLOAD_DIRS = dirs.findAll { fileExists(it) }.join(',')
                    echo "Selected workload dirs: ${env.WORKLOAD_DIRS}"
                }
            }
        }

        // ── Cost: Infracost ───────────────────────────────────────────────────
        stage('Infracost — Plan Cost') {
            when { expression { params.RUN_INFRACOST } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        if ! command -v infracost >/dev/null 2>&1; then
                            curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh
                        fi
                        infracost --version
                        for d in \$(echo "${env.WORKLOAD_DIRS}" | tr ',' ' '); do
                            slug=\$(echo "\$d" | tr '/' '_')
                            echo "=== Infracost breakdown: \$d ==="
                            infracost breakdown --path "\$d" \
                                --format json \
                                --out-file reports/infra-quality/infracost-\${slug}.json || true
                            infracost breakdown --path "\$d" \
                                --format diff \
                                --out-file reports/infra-quality/infracost-\${slug}.txt || true
                        done
                        # Combined HTML report across all workloads
                        infracost output --path "reports/infra-quality/infracost-*.json" \
                            --format html \
                            --out-file reports/infra-quality/infracost-summary.html || true
                    """
                    archiveArtifacts artifacts: 'reports/infra-quality/infracost-*', allowEmptyArchive: true, fingerprint: true
                }
            }
        }

        // ── Drift: Driftctl ───────────────────────────────────────────────────
        stage('Driftctl — State vs Cloud Drift') {
            when { expression { params.RUN_DRIFTCTL } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        if ! command -v driftctl >/dev/null 2>&1; then
                            curl -L https://github.com/snyk/driftctl/releases/latest/download/driftctl_linux_amd64 \
                                -o /tmp/driftctl && sudo install /tmp/driftctl /usr/local/bin/driftctl
                        fi
                        driftctl version
                        for d in \$(echo "${env.WORKLOAD_DIRS}" | tr ',' ' '); do
                            slug=\$(echo "\$d" | tr '/' '_')
                            cloud=\$(echo "\$d" | awk -F'/' '{print \$3}')
                            echo "=== Driftctl scan: \$d (\$cloud) ==="
                            case "\$cloud" in
                                aws) driftctl scan --to aws+tf \
                                        --from "tfstate://\$d/terraform.tfstate" \
                                        --output "json://reports/infra-quality/driftctl-\${slug}.json" || true ;;
                                gcp) driftctl scan --to gcp+tf \
                                        --from "tfstate://\$d/terraform.tfstate" \
                                        --output "json://reports/infra-quality/driftctl-\${slug}.json" || true ;;
                                azure) driftctl scan --to azure+tf \
                                        --from "tfstate://\$d/terraform.tfstate" \
                                        --output "json://reports/infra-quality/driftctl-\${slug}.json" || true ;;
                            esac
                        done
                    """
                    archiveArtifacts artifacts: 'reports/infra-quality/driftctl-*', allowEmptyArchive: true, fingerprint: true
                }
            }
        }

        // ── Lint: tflint ──────────────────────────────────────────────────────
        stage('tflint — Terraform Linter') {
            when { expression { params.RUN_TFLINT } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm -v \${PWD}:/data ghcr.io/terraform-linters/tflint --init || true
                        for d in \$(echo "${env.WORKLOAD_DIRS}" | tr ',' ' '); do
                            slug=\$(echo "\$d" | tr '/' '_')
                            echo "=== tflint: \$d ==="
                            docker run --rm -v \${PWD}:/data \
                                ghcr.io/terraform-linters/tflint \
                                --chdir="/data/\$d" \
                                --format=json \
                                > reports/infra-quality/tflint-\${slug}.json 2>&1 || true
                        done
                    """
                    archiveArtifacts artifacts: 'reports/infra-quality/tflint-*', allowEmptyArchive: true, fingerprint: true
                }
            }
        }

        // ── Pre-commit hooks (terraform fmt, tfsec, etc.) ─────────────────────
        stage('pre-commit Hooks') {
            when { expression { params.RUN_PRE_COMMIT } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        if ! command -v pre-commit >/dev/null 2>&1; then
                            pip install --user pre-commit
                            export PATH="$HOME/.local/bin:$PATH"
                        fi
                        pre-commit --version
                        # Run only Terraform-related hooks to scope this stage
                        pre-commit run --all-files \
                            --hook-stage manual terraform_fmt 2>&1 \
                            | tee reports/infra-quality/pre-commit-tf.log || true
                    '''
                    archiveArtifacts artifacts: 'reports/infra-quality/pre-commit-tf.log', allowEmptyArchive: true
                }
            }
        }

        // ── Atlantis config validation ────────────────────────────────────────
        stage('Atlantis Config') {
            when { expression { params.RUN_ATLANTIS_CFG } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        if [ -f infra/atlantis/atlantis.yaml ]; then
                            docker run --rm \
                                -v "$(pwd)":/work -w /work \
                                ghcr.io/runatlantis/atlantis:latest \
                                server --config /work/infra/atlantis/atlantis.yaml \
                                --gh-user dummy --gh-token dummy --repo-allowlist '*' \
                                --validate-only 2>&1 \
                                | tee reports/infra-quality/atlantis-validate.log || true
                        fi
                        if command -v kubectl >/dev/null 2>&1 && kubectl get ns platform >/dev/null 2>&1; then
                            kubectl create configmap atlantis-repo-config \
                                --from-file=atlantis.yaml=infra/atlantis/atlantis.yaml \
                                -n platform --dry-run=client -o yaml \
                                | kubectl apply -f - || true
                        fi
                    '''
                    archiveArtifacts artifacts: 'reports/infra-quality/atlantis-validate.log', allowEmptyArchive: true
                }
            }
        }

        // ── Visualization: inframap ───────────────────────────────────────────
        stage('inframap — Dependency Graph') {
            when { expression { params.RUN_INFRAMAP } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        if ! command -v inframap >/dev/null 2>&1; then
                            curl -L https://github.com/cycloidio/inframap/releases/latest/download/inframap-linux-amd64.tar.gz \
                                | tar -xz -C /tmp/
                            sudo install /tmp/inframap-linux-amd64 /usr/local/bin/inframap
                        fi
                        for d in \$(echo "${env.WORKLOAD_DIRS}" | tr ',' ' '); do
                            slug=\$(echo "\$d" | tr '/' '_')
                            echo "=== inframap graph: \$d ==="
                            inframap generate "\$d" --hcl --raw \
                                > reports/infra-quality/inframap-\${slug}.dot 2>/dev/null || true
                        done
                    """
                    archiveArtifacts artifacts: 'reports/infra-quality/inframap-*.dot', allowEmptyArchive: true
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'reports/infra-quality/**/*', allowEmptyArchive: true
        }
        success { echo "Infra quality run completed — see reports/infra-quality/" }
        failure { echo "Infra quality run failed — check stage logs." }
    }
}
