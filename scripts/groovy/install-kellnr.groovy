def call() {
    sh """
        helm upgrade --install kellnr registry/charts/kellnr \
            --namespace kellnr \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://kellnr-kellnr.kellnr.svc.cluster.local:8080'
    sh "sed -i '/^KELLNR_/d' infra.env || true"
    sh "sed -i '/^KELLNR_URL=/d' infra.env 2>/dev/null || true; echo 'KELLNR_URL=http://kellnr-kellnr.kellnr.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^KELLNR_USER=/d' infra.env 2>/dev/null || true; echo 'KELLNR_USER=admin' >> infra.env" 
    sh "sed -i '/^KELLNR_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'KELLNR_PASSWORD=admin' >> infra.env" 

    echo 'kellnr installed — ${url}'
}

return this
