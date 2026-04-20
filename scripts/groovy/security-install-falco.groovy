def call() {
    // Install Falco with eBPF driver and Falcosidekick for alerting
    sh '''
        helm upgrade --install falco security/falco/charts \
            --namespace falco \
            --create-namespace \
            --set driver.kind=ebpf \
            --set falco.grpc.enabled=true \
            --set falco.grpcOutput.enabled=true \
            --set falco.jsonOutput=true \
            --set falco.jsonIncludeOutputProperty=true \
            --set falco.logLevel=info \
            --set falco.priority=debug \
            --set falco.bufferedOutputs=false \
            --set falco.rulesFiles[0]=/etc/falco/falco_rules.yaml \
            --set falco.rulesFiles[1]=/etc/falco/falco_rules.local.yaml \
            --set falco.rulesFiles[2]=/etc/falco/k8s_audit_rules.yaml \
            --set falco.rulesFiles[3]=/etc/falco/aws_cloudtrail_rules.yaml \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=512Mi \
            --set resources.limits.cpu=1000m \
            --set resources.limits.memory=1024Mi \
            --set tolerations[0].effect=NoSchedule \
            --set tolerations[0].key=node-role.kubernetes.io/master \
            --wait --timeout 10m
    '''
    // Install Falcosidekick for forwarding alerts
    sh '''
        helm upgrade --install falcosidekick security/falco/charts/falcosidekick \
            --namespace falco \
            --set config.debug=false \
            --set config.customfields="cluster:shopos" \
            --set config.prometheus.enabled=true \
            --set replicaCount=2 \
            --set webui.enabled=true \
            --set webui.replicaCount=1 \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status daemonset/falco -n falco --timeout=5m"
    sh "sed -i '/^FALCO_/d' infra.env || true"
    sh "sed -i '/^FALCO_GRPC_URL=/d' infra.env 2>/dev/null || true; echo 'FALCO_GRPC_URL=http://falco.falco.svc.cluster.local:5060' >> infra.env" 
    sh "sed -i '/^FALCO_SIDEKICK_URL=/d' infra.env 2>/dev/null || true; echo 'FALCO_SIDEKICK_URL=http://falcosidekick.falco.svc.cluster.local:2801' >> infra.env" 
    sh "sed -i '/^FALCO_SIDEKICK_UI_URL=/d' infra.env 2>/dev/null || true; echo 'FALCO_SIDEKICK_UI_URL=http://falcosidekick-ui.falco.svc.cluster.local:2802' >> infra.env" 
    echo 'Falco installed — eBPF driver with k8s_audit rules, Falcosidekick, and Web UI'
}
return this
