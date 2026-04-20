def call() {
    sh """
        helm upgrade --install geminabox registry/charts/geminabox \
            --namespace geminabox \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://geminabox-geminabox.geminabox.svc.cluster.local:9292'
    sh "sed -i '/^GEMINABOX_/d' infra.env || true"
    sh "sed -i '/^GEMINABOX_URL=/d' infra.env 2>/dev/null || true; echo 'GEMINABOX_URL=http://geminabox-geminabox.geminabox.svc.cluster.local:9292' >> infra.env" 

    echo 'geminabox installed — ${url}'
}

return this
