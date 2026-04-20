def call() {
    sh '''
        echo "=== Configure Traefik ==="

        # Set Traefik as the default IngressClass
        kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: traefik
  annotations:
    ingressclass.kubernetes.io/is-default-class: "true"
spec:
  controller: traefik.io/ingress-controller
EOF

        # Rate-limit middleware for API gateway traffic
        kubectl apply -f - <<EOF
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: ratelimit-api
  namespace: traefik
spec:
  rateLimit:
    average: 100
    burst: 200
EOF

        # HTTPS redirect middleware
        kubectl apply -f - <<EOF
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: redirect-https
  namespace: traefik
spec:
  redirectScheme:
    scheme: https
    permanent: true
EOF

        # Write Traefik dashboard URL to infra.env
        TRAEFIK_LB=$(kubectl get svc traefik -n traefik \
            -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null \
            || kubectl get svc traefik -n traefik \
            -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
        if [ -n "$TRAEFIK_LB" ]; then
            sed -i '/^TRAEFIK_URL=/d' infra.env
            echo "TRAEFIK_URL=http://${TRAEFIK_LB}" >> infra.env
            echo "  TRAEFIK_URL=http://${TRAEFIK_LB}"
        fi

        echo "Traefik IngressClass, rate-limit and redirect middlewares configured."
    '''
}
return this
