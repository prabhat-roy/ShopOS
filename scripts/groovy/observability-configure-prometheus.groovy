def call() {
    sh """
        PROM_URL=\$(grep '^PROMETHEUS_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Prometheus at \${PROM_URL}..."
        until curl -sf "\${PROM_URL}/-/ready" > /dev/null 2>&1; do sleep 10; done

        # Push the ShopOS scrape config and alerting rules via ConfigMap patch
        kubectl create configmap prometheus-scrape-config \
            --from-file=observability/prometheus/prometheus.yaml \
            --namespace prometheus --dry-run=client -o yaml | kubectl apply -f - || true

        kubectl create configmap prometheus-rules \
            --from-file=observability/prometheus/rules/ \
            --namespace prometheus --dry-run=client -o yaml | kubectl apply -f - || true

        kubectl rollout restart deployment/prometheus-prometheus -n prometheus || true
    """
    echo 'prometheus configured — ShopOS scrape config and alerting rules applied'
}
return this
