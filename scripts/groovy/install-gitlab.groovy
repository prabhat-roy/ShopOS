def call() {
    sh """
        helm upgrade --install gitlab registry/charts/gitlab \
            --namespace gitlab \
            --create-namespace \
            --wait --timeout 15m
    """

    def url = 'http://gitlab-gitlab.gitlab.svc.cluster.local:80'
    sh "sed -i '/^GITLAB_/d' infra.env || true"
    sh "echo 'GITLAB_URL=http://gitlab-gitlab.gitlab.svc.cluster.local:80' >> infra.env"
    sh "echo 'GITLAB_USER=root' >> infra.env"
    sh """
        GITLAB_PWD=$(kubectl get secret gitlab-gitlab-initial-root-password \
            -n gitlab -o jsonpath='{.data.password}' | base64 -d 2>/dev/null || echo 'see-gitlab-secret')
        echo "GITLAB_PASSWORD=${GITLAB_PWD}" >> infra.env
    """

    echo 'gitlab installed — ${url}'
}

return this
