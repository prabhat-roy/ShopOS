def call() {
    sh """
        helm upgrade --install onedev registry/charts/onedev \
            --namespace onedev \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://onedev-onedev.onedev.svc.cluster.local:6610'
    sh "sed -i '/^ONEDEV_/d' infra.env || true"
    sh "echo 'ONEDEV_URL=http://onedev-onedev.onedev.svc.cluster.local:6610' >> infra.env"
    sh "echo 'ONEDEV_USER=admin' >> infra.env"

    echo 'onedev installed — ${url}'
}

return this
