def call() {
    sh '''
        helm upgrade --install authelia security/authelia/charts \
            --namespace authelia \
            --create-namespace \
            --set domain=shopos.local \
            --set pod.replicas=2 \
            --set pod.resources.requests.cpu=100m \
            --set pod.resources.requests.memory=128Mi \
            --set configMap.theme=auto \
            --set configMap.default_redirection_url=https://shopos.local \
            --set configMap.session.domain=shopos.local \
            --set configMap.session.expiration=3600 \
            --set configMap.session.inactivity=300 \
            --set configMap.session.redis.enabled=true \
            --set configMap.session.redis.host=redis.redis.svc.cluster.local \
            --set configMap.regulation.max_retries=3 \
            --set configMap.regulation.find_time=2m \
            --set configMap.regulation.ban_time=5m \
            --set configMap.authentication_backend.file.enabled=true \
            --set configMap.authentication_backend.file.path=/config/users_database.yml \
            --set configMap.access_control.default_policy=deny \
            --set configMap.totp.issuer=authelia.shopos.local \
            --set configMap.totp.period=30 \
            --set secret.jwt.value=change-me-jwt-secret \
            --set secret.session.value=change-me-session-secret \
            --set secret.storageEncryptionKey.value=change-me-32chars-encryption-key1 \
            --set metrics.serviceMonitor.enabled=false \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/authelia -n authelia --timeout=5m"
    sh "sed -i '/^AUTHELIA_/d' infra.env || true"
    sh "echo 'AUTHELIA_URL=http://authelia.authelia.svc.cluster.local:9091' >> infra.env"
    echo 'Authelia installed — SSO+2FA proxy, Redis session store, TOTP, default-deny policy'
}
return this
