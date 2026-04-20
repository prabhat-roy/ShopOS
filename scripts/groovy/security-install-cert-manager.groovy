def call() {
    sh '''
        helm upgrade --install cert-manager security/cert-manager/charts \
            --namespace cert-manager \
            --create-namespace \
            --set crds.enabled=true \
            --set replicaCount=2 \
            --set webhook.replicaCount=2 \
            --set cainjector.replicaCount=2 \
            --set global.leaderElection.namespace=cert-manager \
            --set prometheus.enabled=true \
            --set prometheus.servicemonitor.enabled=false \
            --set resources.requests.cpu=10m \
            --set resources.requests.memory=32Mi \
            --set webhook.resources.requests.cpu=10m \
            --set webhook.resources.requests.memory=32Mi \
            --set cainjector.resources.requests.cpu=10m \
            --set cainjector.resources.requests.memory=32Mi \
            --set featureGates="" \
            --set dns01RecursiveNameserversOnly=false \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/cert-manager -n cert-manager --timeout=5m"
    // Create a self-signed ClusterIssuer for cluster-internal TLS
    sh '''
        kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: ca-issuer
spec:
  ca:
    secretName: ca-key-pair
EOF
    '''
    sh "sed -i '/^CERT_MANAGER_/d' infra.env || true"
    sh "echo 'CERT_MANAGER_URL=http://cert-manager.cert-manager.svc.cluster.local:9402' >> infra.env"
    echo 'cert-manager installed — CRDs, 2 replicas, Prometheus metrics, selfsigned ClusterIssuer'
}
return this
