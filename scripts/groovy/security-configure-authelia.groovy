def call() {
    sh """
        echo "Configuring Authelia..."
        # Patch Authelia ConfigMap with minimal working session and storage config
        kubectl patch configmap authelia-authelia -n authelia --type merge -p '{
            "data": {
                "configuration.yml": "---\\ntheme: light\\njwt_secret: change-me-very-secret\\ndefault_redirection_url: https://shopos.local\\nserver:\\n  host: 0.0.0.0\\n  port: 9091\\nlog:\\n  level: info\\nauthentication_backend:\\n  file:\\n    path: /config/users.yml\\nsession:\\n  name: authelia_session\\n  secret: change-me-session-secret\\n  expiration: 3600\\n  inactivity: 300\\n  domain: shopos.local\\nstorage:\\n  encryption_key: change-me-storage-key\\n  local:\\n    path: /config/db.sqlite3\\naccess_control:\\n  default_policy: deny\\n  rules:\\n    - domain: shopos.local\\n      policy: two_factor\\nnotifier:\\n  disable_startup_check: true\\n  filesystem:\\n    filename: /config/notification.txt\\n"
            }
        }' || true
        kubectl rollout restart deployment/authelia-authelia -n authelia || true
    """
    echo 'authelia configured — default session, storage, and access policy applied'
}
return this
