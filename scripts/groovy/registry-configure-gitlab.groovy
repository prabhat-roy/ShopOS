def call() {
    sh '''
        echo "=== Configure GitLab ==="

        kubectl rollout status deploy/gitlab-webservice-default -n gitlab --timeout=300s || true

        GITLAB_IP=$(kubectl get svc gitlab-webservice-default -n gitlab \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "gitlab-webservice-default.gitlab.svc.cluster.local")
        GITLAB_URL="http://${GITLAB_IP}"

        # Retrieve initial root password
        ROOT_PASS=$(kubectl get secret gitlab-gitlab-initial-root-password -n gitlab \
            -o jsonpath='{.data.password}' 2>/dev/null | base64 -d || echo "")

        sed -i '/^GITLAB_URL=/d; /^GITLAB_ROOT_PASS=/d' infra.env
        echo "GITLAB_URL=${GITLAB_URL}" >> infra.env
        [ -n "$ROOT_PASS" ] && echo "GITLAB_ROOT_PASS=${ROOT_PASS}" >> infra.env

        echo "GitLab URL and root password written to infra.env."
    '''
}
return this
