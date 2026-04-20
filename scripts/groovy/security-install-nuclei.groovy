def call() {
    sh '''
        helm upgrade --install nuclei security/nuclei/charts \
            --namespace nuclei \
            --create-namespace \
            --set image.tag=v3.3.4 \
            --set replicaCount=1 \
            --set resources.requests.cpu=500m \
            --set resources.requests.memory=512Mi \
            --set resources.limits.cpu=2000m \
            --set resources.limits.memory=2Gi \
            --set config.serverAddress=0.0.0.0 \
            --set config.serverPort=9090 \
            --set config.serverToken=nucleitoken \
            --set config.templates.update=true \
            --set config.templates.dir=/root/nuclei-templates \
            --set config.severity[0]=critical \
            --set config.severity[1]=high \
            --set config.severity[2]=medium \
            --set config.exclude[0]=ssl \
            --set config.exclude[1]=dns \
            --set config.retryCount=2 \
            --set config.timeout=10 \
            --set config.rateLimit=150 \
            --set persistence.enabled=true \
            --set persistence.size=5Gi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/nuclei -n nuclei --timeout=5m"
    sh "sed -i '/^NUCLEI_/d' infra.env || true"
    sh "echo 'NUCLEI_URL=http://nuclei.nuclei.svc.cluster.local:9090' >> infra.env"
    sh "echo 'NUCLEI_TOKEN=nucleitoken' >> infra.env"
    echo 'Nuclei scanner installed — server mode, auto template updates, crit/high/medium severity'
}
return this
