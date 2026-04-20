def call() {
    sh """
        helm upgrade --install gitea registry/charts/gitea \
            --namespace gitea \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://gitea-gitea.gitea.svc.cluster.local:3000'
    sh "sed -i '/^GITEA_/d' infra.env || true"
    sh "echo 'GITEA_URL=http://gitea-gitea.gitea.svc.cluster.local:3000' >> infra.env"
    sh "echo 'GITEA_USER=gitea_admin' >> infra.env"
    sh "echo 'GITEA_PASSWORD=gitea_admin' >> infra.env"

    echo 'gitea installed — ${url}'
}

return this
