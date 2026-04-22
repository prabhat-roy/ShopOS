def call() {
    sh """
        helm upgrade --install robusta observability/robusta/charts \
            --namespace monitoring \
            --create-namespace \
            --set fullnameOverride=robusta \
            --set env.ROBUSTA_CLUSTER_NAME=shopos-prod \
            --wait --timeout 5m
    """
    sh "sed -i '/^ROBUSTA_/d' infra.env || true"
    sh "echo 'ROBUSTA_URL=http://robusta.monitoring.svc.cluster.local:80' >> infra.env"
    echo 'Robusta installed'
}
return this
