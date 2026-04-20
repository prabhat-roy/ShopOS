def call() {
    sh '''
        helm upgrade --install clair security/clair/charts \
            --namespace clair \
            --create-namespace \
            --set image.tag=4.7.4 \
            --set replicas=2 \
            --set resources.requests.cpu=200m \
            --set resources.requests.memory=512Mi \
            --set resources.limits.cpu=1000m \
            --set resources.limits.memory=2Gi \
            --set config.mode=combo \
            --set config.http_listen_addr=0.0.0.0:6060 \
            --set config.introspection_addr=0.0.0.0:8089 \
            --set config.log_level=info \
            --set config.matcher.period=6h \
            --set config.updaters.sets[0]=alpine \
            --set config.updaters.sets[1]=aws \
            --set config.updaters.sets[2]=debian \
            --set config.updaters.sets[3]=oracle \
            --set config.updaters.sets[4]=photon \
            --set config.updaters.sets[5]=pyupio \
            --set config.updaters.sets[6]=rhel \
            --set config.updaters.sets[7]=suse \
            --set config.updaters.sets[8]=ubuntu \
            --set postgresql.enabled=true \
            --set postgresql.auth.database=clair \
            --set postgresql.auth.username=clair \
            --set postgresql.auth.password=clair \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status deployment/clair -n clair --timeout=5m"
    sh "sed -i '/^CLAIR_/d' infra.env || true"
    sh "sed -i '/^CLAIR_URL=/d' infra.env 2>/dev/null || true; echo 'CLAIR_URL=http://clair.clair.svc.cluster.local:6060' >> infra.env" 
    sh "sed -i '/^CLAIR_INTROSPECTION_URL=/d' infra.env 2>/dev/null || true; echo 'CLAIR_INTROSPECTION_URL=http://clair.clair.svc.cluster.local:8089' >> infra.env" 
    echo 'Clair installed — combo mode, 6h updater interval for all Linux distros, PostgreSQL'
}
return this
