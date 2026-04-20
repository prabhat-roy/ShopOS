def call() {
    sh '''
        echo "=== Configure Istio ==="

        # Enable strict mTLS across all namespaces
        kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: istio-system
spec:
  mtls:
    mode: STRICT
EOF

        # Enable sidecar injection for all ShopOS namespaces
        for ns in commerce platform identity catalog supply-chain financial customer-experience \\
                  communications content analytics-ai b2b integrations affiliate; do
            kubectl label namespace $ns istio-injection=enabled --overwrite 2>/dev/null || true
        done

        # Default destination rule — ISTIO_MUTUAL TLS mode for all services
        kubectl apply -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: default
  namespace: istio-system
spec:
  host: "*.local"
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
EOF

        # Write Istio ingress gateway URL to infra.env
        ISTIO_LB=$(kubectl get svc istio-ingressgateway -n istio-system \
            -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null \
            || kubectl get svc istio-ingressgateway -n istio-system \
            -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
        if [ -n "$ISTIO_LB" ]; then
            sed -i '/^ISTIO_GATEWAY_URL=/d' infra.env
            echo "ISTIO_GATEWAY_URL=http://${ISTIO_LB}" >> infra.env
            echo "  ISTIO_GATEWAY_URL=http://${ISTIO_LB}"
        fi

        echo "Istio mTLS and sidecar injection configured."
    '''
}
return this
