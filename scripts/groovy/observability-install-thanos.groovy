def call() {
    sh """
        helm upgrade --install thanos observability/thanos/charts             --namespace thanos             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^THANOS_/d' infra.env || true"
    sh "echo 'THANOS_URL=http://thanos-thanos.thanos.svc.cluster.local:10902' >> infra.env"
    echo 'thanos installed'
}
return this
