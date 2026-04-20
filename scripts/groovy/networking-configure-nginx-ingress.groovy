def call() {
    sh '''
        echo "=== Configure NGINX Ingress ==="

        # Set NGINX as the default IngressClass
        kubectl annotate ingressclass nginx ingressclass.kubernetes.io/is-default-class=true --overwrite 2>/dev/null || true

        # Patch ConfigMap — enable real IP, set proxy buffer sizes
        kubectl patch configmap nginx-configuration -n nginx-ingress --type merge -p '{
          "data": {
            "use-forwarded-headers": "true",
            "proxy-body-size": "50m",
            "proxy-buffer-size": "16k",
            "keep-alive-requests": "10000"
          }
        }' 2>/dev/null || true

        echo "NGINX Ingress default IngressClass and ConfigMap configured."
    '''
}
return this
