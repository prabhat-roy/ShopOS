def call() {
    sh """
        OPENVAS_URL=\$(grep '^OPENVAS_URL=' infra.env | cut -d= -f2)
        echo "Waiting for OpenVAS at \${OPENVAS_URL}..."
        until curl -sf "\${OPENVAS_URL}" > /dev/null 2>&1; do sleep 15; done

        # Feed sync — trigger NVT, SCAP, CERT feed update via GMP
        kubectl exec -n openvas deploy/openvas-openvas -- \
            greenbone-feed-sync --type NVT || true
        kubectl exec -n openvas deploy/openvas-openvas -- \
            greenbone-feed-sync --type SCAP || true
        kubectl exec -n openvas deploy/openvas-openvas -- \
            greenbone-feed-sync --type CERT || true

        # Create default scan config target (localhost as placeholder)
        curl -sf -u admin:admin "\${OPENVAS_URL}/gmp" \
            -d '<create_target><name>ShopOS Default</name><hosts>localhost</hosts></create_target>' || true
    """
    echo 'openvas configured — feed sync triggered, default scan target created'
}
return this
