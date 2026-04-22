def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        # Clear any pending/failed helm state before installing
        STATUS=\$(helm status pulsar -n pulsar 2>/dev/null | grep STATUS | awk '{print \$2}')
        if [ "\$STATUS" = "pending-install" ] || [ "\$STATUS" = "pending-upgrade" ] || [ "\$STATUS" = "failed" ]; then
            echo "Clearing stuck helm release: \$STATUS"
            helm uninstall pulsar -n pulsar 2>/dev/null || true
            kubectl delete pvc data-pulsar-0 -n pulsar 2>/dev/null || true
            kubectl delete pod --all -n pulsar --force 2>/dev/null || true
        fi

        helm upgrade --install pulsar messaging/pulsar/charts \
            --namespace pulsar \
            --create-namespace \
            --set fullnameOverride=pulsar \
            --set persistence.storageClass=${sc} \
            --wait --timeout 10m
    """
    sh "sed -i '/^PULSAR_/d' infra.env || true"
    sh "echo 'PULSAR_URL=pulsar://pulsar.pulsar.svc.cluster.local:6650' >> infra.env"
    sh "echo 'PULSAR_HTTP_URL=http://pulsar.pulsar.svc.cluster.local:8080' >> infra.env"
    echo 'pulsar installed'
}
return this
