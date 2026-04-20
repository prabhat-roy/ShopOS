def call() {
    sh """
        helm upgrade --install memphis messaging/memphis/charts \
            --namespace memphis \
            --create-namespace \
            --set env.ROOT_PASSWORD=memphis \
            --wait --timeout 5m
    """
    sh "sed -i '/^MEMPHIS_/d' infra.env || true"
    sh "echo 'MEMPHIS_URL=memphis-memphis.memphis.svc.cluster.local:6666' >> infra.env"
    sh "echo 'MEMPHIS_HTTP_URL=http://memphis-memphis.memphis.svc.cluster.local:9000' >> infra.env"
    sh "echo 'MEMPHIS_USER=root' >> infra.env"
    sh "echo 'MEMPHIS_PASSWORD=memphis' >> infra.env"
    echo 'memphis installed'
}
return this
