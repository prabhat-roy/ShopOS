def call() {
    sh """
        helm upgrade --install glitchtip observability/glitchtip/charts             --namespace glitchtip             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^GLITCHTIP_/d' infra.env || true"
    sh "sed -i '/^GLITCHTIP_URL=/d' infra.env 2>/dev/null || true; echo 'GLITCHTIP_URL=http://glitchtip-glitchtip.glitchtip.svc.cluster.local:8000' >> infra.env" 
    echo 'glitchtip installed'
}
return this
