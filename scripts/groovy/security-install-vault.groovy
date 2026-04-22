def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install vault security/vault/charts \
            --namespace vault \
            --create-namespace \
            --set server.ha.enabled=true \
            --set server.ha.replicas=3 \
            --set server.ha.raft.enabled=true \
            --set server.ha.raft.setNodeId=true \
            --set server.image.tag=1.17.5 \
            --set server.resources.requests.cpu=250m \
            --set server.resources.requests.memory=256Mi \
            --set server.resources.limits.cpu=500m \
            --set server.resources.limits.memory=512Mi \
            --set server.dataStorage.enabled=true \
            --set server.dataStorage.size=10Gi \
            --set server.dataStorage.storageClass=${sc} \
            --set server.auditStorage.enabled=true \
            --set server.auditStorage.size=5Gi \
            --set server.auditStorage.storageClass=${sc} \
            --set server.serviceAccount.create=true \
            --set server.serviceAccount.name=vault \
            --set server.extraEnvironmentVars.VAULT_CACERT=/vault/userconfig/vault-ha-tls/ca.crt \
            --set server.readinessProbe.enabled=true \
            --set server.readinessProbe.path=/v1/sys/health?standbyok=true \
            --set server.livenessProbe.enabled=true \
            --set server.livenessProbe.path=/v1/sys/health?standbyok=true \
            --set 'server.affinity=' \
            --set injector.enabled=true \
            --set injector.replicas=2 \
            --set injector.metrics.enabled=true \
            --set ui.enabled=true \
            --set ui.serviceType=ClusterIP \
            --set csi.enabled=true \
            --wait --timeout 10m
    """
    sh "kubectl rollout status statefulset/vault -n vault --timeout=5m"
    sh "sed -i '/^VAULT_/d' infra.env || true"
    sh "sed -i '/^VAULT_URL=/d' infra.env 2>/dev/null || true; echo 'VAULT_URL=http://vault.vault.svc.cluster.local:8200' >> infra.env" 
    sh "sed -i '/^VAULT_INJECTOR_URL=/d' infra.env 2>/dev/null || true; echo 'VAULT_INJECTOR_URL=http://vault-agent-injector.vault.svc.cluster.local:8080' >> infra.env" 
    echo 'Vault installed — HA Raft mode with 3 replicas, agent injector, CSI driver, and UI'
}
return this
