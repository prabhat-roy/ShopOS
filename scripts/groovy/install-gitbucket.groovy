def call() {
    sh """
        helm upgrade --install gitbucket registry/charts/gitbucket \
            --namespace gitbucket \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://gitbucket-gitbucket.gitbucket.svc.cluster.local:8080'
    sh "sed -i '/^GITBUCKET_/d' infra.env || true"
    sh "sed -i '/^GITBUCKET_URL=/d' infra.env 2>/dev/null || true; echo 'GITBUCKET_URL=http://gitbucket-gitbucket.gitbucket.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^GITBUCKET_USER=/d' infra.env 2>/dev/null || true; echo 'GITBUCKET_USER=root' >> infra.env" 
    sh "sed -i '/^GITBUCKET_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'GITBUCKET_PASSWORD=root' >> infra.env" 

    echo 'gitbucket installed — ${url}'
}

return this
