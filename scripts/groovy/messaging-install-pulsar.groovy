def call() {
    sh """
        helm upgrade --install pulsar messaging/pulsar/charts \
            --namespace pulsar \
            --create-namespace \
            --set env.PULSAR_MEM="-Xms512m -Xmx512m -XX:MaxDirectMemorySize=256m" \
            --wait --timeout 10m
    """
    sh "sed -i '/^PULSAR_/d' infra.env || true"
    sh "sed -i '/^PULSAR_URL=/d' infra.env 2>/dev/null || true; echo 'PULSAR_URL=pulsar://pulsar-pulsar.pulsar.svc.cluster.local:6650' >> infra.env" 
    sh "sed -i '/^PULSAR_HTTP_URL=/d' infra.env 2>/dev/null || true; echo 'PULSAR_HTTP_URL=http://pulsar-pulsar.pulsar.svc.cluster.local:8080' >> infra.env" 
    echo 'pulsar installed'
}
return this
