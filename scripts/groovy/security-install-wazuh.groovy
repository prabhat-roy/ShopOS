def call() {
    sh '''
        helm upgrade --install wazuh security/wazuh/charts \
            --namespace wazuh \
            --create-namespace \
            --set wazuh-manager.replicaCount=1 \
            --set wazuh-manager.resources.requests.cpu=500m \
            --set wazuh-manager.resources.requests.memory=512Mi \
            --set wazuh-manager.resources.limits.cpu=2000m \
            --set wazuh-manager.resources.limits.memory=2Gi \
            --set wazuh-manager.storage.size=10Gi \
            --set wazuh-dashboard.replicas=1 \
            --set wazuh-dashboard.resources.requests.cpu=100m \
            --set wazuh-dashboard.resources.requests.memory=512Mi \
            --set wazuh-indexer.replicas=1 \
            --set wazuh-indexer.resources.requests.cpu=500m \
            --set wazuh-indexer.resources.requests.memory=1Gi \
            --set wazuh-indexer.storage.size=20Gi \
            --set wazuh-indexer.config.node.name=node-1 \
            --set wazuh-indexer.config.cluster.initial_master_nodes=node-1 \
            --set global.indexer.password=SecretPassword \
            --set global.api.password=SecretPassword \
            --wait --timeout 20m
    '''
    sh "kubectl rollout status statefulset/wazuh-manager -n wazuh --timeout=10m"
    sh "sed -i '/^WAZUH_/d' infra.env || true"
    sh "sed -i '/^WAZUH_URL=/d' infra.env 2>/dev/null || true; echo 'WAZUH_URL=https://wazuh-manager.wazuh.svc.cluster.local:55000' >> infra.env" 
    sh "sed -i '/^WAZUH_DASHBOARD_URL=/d' infra.env 2>/dev/null || true; echo 'WAZUH_DASHBOARD_URL=https://wazuh-dashboard.wazuh.svc.cluster.local:443' >> infra.env" 
    sh "sed -i '/^WAZUH_USER=/d' infra.env 2>/dev/null || true; echo 'WAZUH_USER=admin' >> infra.env" 
    sh "sed -i '/^WAZUH_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'WAZUH_PASSWORD=SecretPassword' >> infra.env" 
    echo 'Wazuh XDR/SIEM installed — manager, indexer, dashboard with persistent storage'
}
return this
