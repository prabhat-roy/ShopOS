def call() {
    sh """
        helm upgrade --install logstash observability/logstash/charts             --namespace logstash             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^LOGSTASH_/d' infra.env || true"
    sh "sed -i '/^LOGSTASH_URL=/d' infra.env 2>/dev/null || true; echo 'LOGSTASH_URL=http://logstash-logstash.logstash.svc.cluster.local:5044' >> infra.env" 
    echo 'logstash installed'
}
return this
