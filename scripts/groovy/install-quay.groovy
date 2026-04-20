def call() {
    sh """
        helm upgrade --install quay registry/charts/quay \
            --namespace quay \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://quay-quay.quay.svc.cluster.local:8080'
    sh "sed -i '/^QUAY_/d' infra.env || true"
    sh "sed -i '/^QUAY_URL=/d' infra.env 2>/dev/null || true; echo 'QUAY_URL=http://quay-quay.quay.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^QUAY_USER=/d' infra.env 2>/dev/null || true; echo 'QUAY_USER=quay' >> infra.env" 
    sh "sed -i '/^QUAY_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'QUAY_PASSWORD=password' >> infra.env" 

    echo 'quay installed — ${url}'
}

return this
