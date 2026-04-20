def call() {
    def svc      = env.TEST_SERVICE
    def ns       = env.TEST_NAMESPACE
    def duration = env.CHAOS_DURATION ?: '2m'
    def report   = 'reports/chaos/chaos-mesh.json'

    sh """
        echo "=== Chaos Mesh: pod-kill on ${svc} ==="

        cat <<EOF | kubectl apply -f - 2>/dev/null || true
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: postdeploy-pod-kill-${svc}
  namespace: ${ns}
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces: ["${ns}"]
    labelSelectors:
      app.kubernetes.io/name: "${svc}"
  duration: "${duration}"
EOF

        echo "Chaos experiment applied — waiting ${duration} for recovery..."
        sleep \$(echo "${duration}" | sed 's/m/*60/' | sed 's/s//' | bc 2>/dev/null || echo 120)

        kubectl delete podchaos postdeploy-pod-kill-${svc} -n ${ns} --ignore-not-found 2>/dev/null || true

        # Wait for pod to recover
        kubectl rollout status deployment/${svc} -n ${ns} --timeout=120s || true

        READY=\$(kubectl get deployment ${svc} -n ${ns} \
            -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo 0)
        DESIRED=\$(kubectl get deployment ${svc} -n ${ns} \
            -o jsonpath='{.spec.replicas}' 2>/dev/null || echo 1)

        if [ "\${READY}" -ge "\${DESIRED}" ]; then
            VERDICT="PASS"
        else
            VERDICT="FAIL"
            echo "WARNING: ${svc} not fully recovered after pod-kill (ready=\${READY} desired=\${DESIRED})"
        fi

        mkdir -p reports/chaos
        cat > ${report} <<REPORT
{
  "tool": "chaos-mesh",
  "service": "${svc}",
  "experiment": "pod-kill",
  "duration": "${duration}",
  "ready_replicas": \${READY},
  "desired_replicas": \${DESIRED},
  "verdict": "\${VERDICT}"
}
REPORT

        echo "Chaos Mesh pod-kill result: \${VERDICT}"
    """
}
return this
