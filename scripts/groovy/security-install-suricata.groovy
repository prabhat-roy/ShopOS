def call() {
    sh '''
        helm upgrade --install suricata security/suricata/charts \
            --namespace suricata \
            --create-namespace \
            --set suricata.config.homeNet=192.168.0.0/16,10.0.0.0/8,172.16.0.0/12 \
            --set suricata.config.defaultLogDir=/var/log/suricata \
            --set suricata.config.af-packet[0].interface=eth0 \
            --set suricata.config.stats.enabled=true \
            --set suricata.config.stats.interval=8 \
            --set suricata.config.outputs[0].eve-log.enabled=true \
            --set suricata.config.outputs[0].eve-log.filename=eve.json \
            --set suricata.config.outputs[0].eve-log.types[0].alert.enabled=true \
            --set suricata.config.outputs[0].eve-log.types[0].alert.metadata=true \
            --set suricata.config.outputs[0].eve-log.types[1].flow.enabled=false \
            --set suricata.config.outputs[0].eve-log.types[2].http.enabled=true \
            --set suricata.config.outputs[0].eve-log.types[3].dns.enabled=true \
            --set suricata.rules.enabled=true \
            --set suricata.rules.emergingthreats.enabled=true \
            --set resources.requests.cpu=200m \
            --set resources.requests.memory=512Mi \
            --set resources.limits.cpu=2000m \
            --set resources.limits.memory=2Gi \
            --set metrics.enabled=true \
            --set metrics.port=8080 \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status daemonset/suricata -n suricata --timeout=5m"
    sh "sed -i '/^SURICATA_/d' infra.env || true"
    sh "sed -i '/^SURICATA_METRICS_URL=/d' infra.env 2>/dev/null || true; echo 'SURICATA_METRICS_URL=http://suricata.suricata.svc.cluster.local:8080/metrics' >> infra.env" 
    echo 'Suricata IDS/IPS installed — DaemonSet, EVE JSON logs, Emerging Threats rules, Prometheus metrics'
}
return this
