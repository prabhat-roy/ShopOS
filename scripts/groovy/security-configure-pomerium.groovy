def call() {
    sh """
        echo "Configuring Pomerium..."
        # Patch Pomerium ConfigMap with default identity provider config
        kubectl patch configmap pomerium-pomerium -n pomerium --type merge -p '{
            "data": {
                "config.yaml": "authenticate:\\n  callback_path: /oauth2/callback\\ncookie:\\n  name: _pomerium\\n  secret: change-me-cookie-secret\\nshared_secret: change-me-shared-secret\\nidp_provider: oidc\\nidp_provider_url: http://keycloak-keycloak.keycloak.svc.cluster.local:8080/realms/shopos\\nidp_client_id: shopos-app\\nidp_client_secret: change-me\\npolicy:\\n  - from: https://app.shopos.local\\n    to: http://web-bff.platform.svc.cluster.local:8081\\n    allowed_domains:\\n      - shopos.local\\n"
            }
        }' || true
        kubectl rollout restart deployment/pomerium-pomerium -n pomerium || true
    """
    echo 'pomerium configured — default OIDC config pointing to Keycloak shopos realm'
}
return this
