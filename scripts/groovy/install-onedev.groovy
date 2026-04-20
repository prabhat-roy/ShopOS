def call() {
    sh """
        helm upgrade --install onedev registry/charts/onedev \
            --namespace onedev \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://onedev-onedev.onedev.svc.cluster.local:6610'
    sh "sed -i '/^ONEDEV_/d' infra.env || true"
    sh "sed -i '/^ONEDEV_URL=/d' infra.env 2>/dev/null || true; echo 'ONEDEV_URL=http://onedev-onedev.onedev.svc.cluster.local:6610' >> infra.env" 
    sh "sed -i '/^ONEDEV_USER=/d' infra.env 2>/dev/null || true; echo 'ONEDEV_USER=admin' >> infra.env" 

    echo 'onedev installed — ${url}'
}

return this
