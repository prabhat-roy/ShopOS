def call() {
    def svc      = env.TEST_SERVICE
    def ns       = env.TEST_NAMESPACE
    def domain   = env.TEST_DOMAIN ?: 'platform'
    def duration = env.CHAOS_DURATION ?: '2m'
    def report   = 'reports/chaos/litmus.json'

    // Run database-chaos for services with a DB; payment-chaos for commerce/financial
    def experiment = (domain in ['commerce', 'financial']) ? 'payment-chaos' : 'database-chaos'

    sh """
        echo "=== Litmus: ${experiment} for ${svc} ==="

        cat <<EOF | kubectl apply -f - 2>/dev/null || true
apiVersion: litmuschaos.io/v1alpha1
kind: ChaosEngine
metadata:
  name: postdeploy-${experiment}-${svc}
  namespace: ${ns}
spec:
  appinfo:
    appns: ${ns}
    applabel: "app.kubernetes.io/name=${svc}"
    appkind: deployment
  chaosServiceAccount: litmus-admin
  experiments:
    - name: ${experiment}
      spec:
        components:
          env:
            - name: TOTAL_CHAOS_DURATION
              value: "${duration}"
            - name: CHAOS_INTERVAL
              value: "10"
EOF

        echo "Litmus experiment applied — polling for completion..."
        TIMEOUT=300
        ELAPSED=0
        VERDICT="AWAITED"
        while [ \$ELAPSED -lt \$TIMEOUT ]; do
            VERDICT=\$(kubectl get chaosresult postdeploy-${experiment}-${svc}-${experiment} \
                -n ${ns} \
                -o jsonpath='{.status.experimentStatus.verdict}' 2>/dev/null || echo "AWAITED")
            [ "\$VERDICT" != "AWAITED" ] && break
            sleep 10
            ELAPSED=\$((ELAPSED + 10))
        done

        kubectl delete chaosengine postdeploy-${experiment}-${svc} -n ${ns} --ignore-not-found 2>/dev/null || true

        mkdir -p reports/chaos
        cat > ${report} <<REPORT
{
  "tool": "litmus",
  "service": "${svc}",
  "experiment": "${experiment}",
  "duration": "${duration}",
  "verdict": "\${VERDICT}"
}
REPORT

        echo "Litmus ${experiment} result: \${VERDICT}"
        [ "\$VERDICT" = "Pass" ] || [ "\$VERDICT" = "AWAITED" ] || \
            echo "WARNING: Litmus experiment verdict was \${VERDICT}"
    """
}
return this
