def call() {
    sh """
        ZAP_URL=\$(grep '^ZAP_URL=' infra.env | cut -d= -f2)
        echo "Waiting for ZAP daemon at \${ZAP_URL}..."
        until curl -sf "\${ZAP_URL}/JSON/core/view/version/" > /dev/null 2>&1; do sleep 10; done

        # Create default context for ShopOS
        CTX_ID=\$(curl -sf "\${ZAP_URL}/JSON/context/action/newContext/?contextName=ShopOS" \
            | grep -o '"contextId":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Set passive scan to scan all requests
        curl -sf "\${ZAP_URL}/JSON/pscan/action/enableAllScanners/" > /dev/null || true

        # Set active scan policy strength to medium
        curl -sf "\${ZAP_URL}/JSON/ascan/action/setOptionAttackStrength/?id=0&attackStrength=MEDIUM" > /dev/null || true
        curl -sf "\${ZAP_URL}/JSON/ascan/action/setOptionAlertThreshold/?id=0&alertThreshold=MEDIUM" > /dev/null || true

        sed -i '/^ZAP_CONTEXT_ID=/d' infra.env || true
        echo "ZAP_CONTEXT_ID=\${CTX_ID}" >> infra.env
    """
    echo 'zap configured — ShopOS context created, scan policies set to medium strength'
}
return this
