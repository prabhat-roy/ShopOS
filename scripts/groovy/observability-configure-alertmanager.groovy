def call() {
    sh """
        AM_URL=\$(grep '^ALERTMANAGER_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Alertmanager at \${AM_URL}..."
        until curl -sf "\${AM_URL}/-/ready" > /dev/null 2>&1; do sleep 10; done

        # Load ShopOS alertmanager config from existing config file
        kubectl create configmap alertmanager-config \
            --from-file=observability/alertmanager/ \
            --namespace alertmanager --dry-run=client -o yaml | kubectl apply -f - || true

        kubectl rollout restart deployment/alertmanager-alertmanager -n alertmanager || true
    """
    echo 'alertmanager configured — ShopOS routing config applied'
}
return this
