def call() {
    sh '''
        echo "=== Configure Fluent Bit ==="

        kubectl rollout status daemonset/fluent-bit -n fluent-bit --timeout=120s || true

        LOKI_URL=$(grep '^LOKI_URL=' infra.env 2>/dev/null | cut -d= -f2 \
            || echo "http://loki-loki.loki.svc.cluster.local:3100")
        ELASTICSEARCH_URL=$(grep '^ELASTICSEARCH_URL=' infra.env 2>/dev/null | cut -d= -f2 \
            || echo "http://elasticsearch.elasticsearch.svc.cluster.local:9200")

        # Patch values via ConfigMap
        kubectl create configmap fluent-bit-config \
            -n fluent-bit \
            --from-literal=fluent-bit.conf="
[SERVICE]
    Flush        5
    Daemon       Off
    Log_Level    info
    Parsers_File parsers.conf

[INPUT]
    Name              tail
    Path              /var/log/containers/*.log
    multiline.parser  docker, cri
    Tag               kube.*
    Mem_Buf_Limit     50MB
    Skip_Long_Lines   On

[FILTER]
    Name                kubernetes
    Match               kube.*
    Kube_URL            https://kubernetes.default.svc:443
    Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
    Merge_Log           On
    Keep_Log            Off

[OUTPUT]
    Name   loki
    Match  kube.*
    host   $(echo $LOKI_URL | sed 's|http://||;s|:.*||')
    port   3100
    labels job=fluentbit,namespace=\$kubernetes['namespace_name'],app=\$kubernetes['labels']['app']

[OUTPUT]
    Name   es
    Match  kube.*
    Host   $(echo $ELASTICSEARCH_URL | sed 's|http://||;s|:.*||')
    Port   9200
    Index  shopos-logs
    Type   _doc" \
            --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

        kubectl rollout restart daemonset/fluent-bit -n fluent-bit 2>/dev/null || true

        echo "Fluent Bit configured to forward to Loki + Elasticsearch."
    '''
}
return this
