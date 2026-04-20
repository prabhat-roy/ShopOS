def call() {
    sh '''
        echo "=== Configure Fluentd ==="

        kubectl rollout status daemonset/fluentd -n fluentd --timeout=120s || true

        # Patch Fluentd ConfigMap to forward logs to both Elasticsearch and Loki
        ELASTICSEARCH_URL=$(grep '^ELASTICSEARCH_URL=' infra.env 2>/dev/null | cut -d= -f2 \
            || echo "http://elasticsearch.elasticsearch.svc.cluster.local:9200")
        LOKI_URL=$(grep '^LOKI_URL=' infra.env 2>/dev/null | cut -d= -f2 \
            || echo "http://loki-loki.loki.svc.cluster.local:3100")

        kubectl create configmap fluentd-config \
            -n fluentd \
            --from-literal=fluentd.conf="
<source>
  @type tail
  path /var/log/containers/*.log
  pos_file /var/log/fluentd-containers.log.pos
  tag kubernetes.*
  read_from_head true
  <parse>
    @type json
    time_format %Y-%m-%dT%H:%M:%S.%NZ
  </parse>
</source>
<match kubernetes.**>
  @type copy
  <store>
    @type elasticsearch
    host ${ELASTICSEARCH_URL}
    index_name shopos-logs
    include_timestamp true
  </store>
  <store>
    @type loki
    url ${LOKI_URL}
    <label>
      app \$.kubernetes.labels.app
      namespace \$.kubernetes.namespace_name
    </label>
  </store>
</match>" \
            --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

        kubectl rollout restart daemonset/fluentd -n fluentd 2>/dev/null || true

        echo "Fluentd configured to forward to Elasticsearch + Loki."
    '''
}
return this
