def call(String namespace) {
    echo "=== Applying K8s enhancements: ${namespace} ==="

    // Discover workloads
    def deploys = sh(
        script: "kubectl get deployments -n ${namespace} -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true",
        returnStdout: true
    ).trim()

    def statefulsets = sh(
        script: "kubectl get statefulsets -n ${namespace} -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true",
        returnStdout: true
    ).trim()

    def deployList     = deploys     ? deploys.split()     as List : []
    def statefulList   = statefulsets ? statefulsets.split() as List : []
    def allWorkloads   = deployList + statefulList

    if (allWorkloads.isEmpty()) {
        echo "  No deployments or statefulsets found in ${namespace} — skipping"
        return
    }

    def kedaInstalled = sh(
        script: 'kubectl get crd scaledobjects.keda.sh >/dev/null 2>&1 && echo yes || echo no',
        returnStdout: true
    ).trim() == 'yes'

    // ── HPA / KEDA ScaledObject for Deployments ───────────────────────────────
    for (d in deployList) {
        if (kedaInstalled) {
            writeFile file: "/tmp/keda-so-${d}.yaml", text: """\
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: ${d}-scaledobject
  namespace: ${namespace}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ${d}
  minReplicaCount: 1
  maxReplicaCount: 5
  triggers:
  - type: cpu
    metricType: Utilization
    metadata:
      value: "70"
  - type: memory
    metricType: Utilization
    metadata:
      value: "80"
"""
            sh "kubectl apply -f /tmp/keda-so-${d}.yaml && rm -f /tmp/keda-so-${d}.yaml"
            echo "  KEDA ScaledObject → ${d}"
        } else {
            writeFile file: "/tmp/hpa-${d}.yaml", text: """\
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ${d}-hpa
  namespace: ${namespace}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ${d}
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
"""
            sh "kubectl apply -f /tmp/hpa-${d}.yaml && rm -f /tmp/hpa-${d}.yaml"
            echo "  HPA → ${d}"
        }
    }

    // ── KEDA ScaledObject for StatefulSets (if KEDA available) ───────────────
    for (ss in statefulList) {
        if (kedaInstalled) {
            writeFile file: "/tmp/keda-so-${ss}.yaml", text: """\
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: ${ss}-scaledobject
  namespace: ${namespace}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: ${ss}
  minReplicaCount: 1
  maxReplicaCount: 3
  triggers:
  - type: cpu
    metricType: Utilization
    metadata:
      value: "70"
"""
            sh "kubectl apply -f /tmp/keda-so-${ss}.yaml && rm -f /tmp/keda-so-${ss}.yaml"
            echo "  KEDA ScaledObject (StatefulSet) → ${ss}"
        }
    }

    // ── PodDisruptionBudget for all workloads ─────────────────────────────────
    for (w in allWorkloads) {
        def kind = deployList.contains(w) ? 'Deployment' : 'StatefulSet'
        def matchLabels = sh(
            script: """kubectl get ${kind.toLowerCase()} ${w} -n ${namespace} \
                -o go-template='{{range \$k,\$v := .spec.selector.matchLabels}}      {{\$k}}: {{\$v}}{{"\\n"}}{{end}}' \
                2>/dev/null || true""",
            returnStdout: true
        ).trim()
        if (!matchLabels) matchLabels = "      app: ${w}"

        writeFile file: "/tmp/pdb-${w}.yaml", text: """\
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: ${w}-pdb
  namespace: ${namespace}
spec:
  minAvailable: 1
  selector:
    matchLabels:
${matchLabels}
"""
        sh "kubectl apply -f /tmp/pdb-${w}.yaml && rm -f /tmp/pdb-${w}.yaml"
        echo "  PDB → ${w}"
    }

    // ── Vertical Pod Autoscaler (if VPA CRD is available) ────────────────────
    def vpaInstalled = sh(
        script: 'kubectl get crd verticalpodautoscalers.autoscaling.k8s.io >/dev/null 2>&1 && echo yes || echo no',
        returnStdout: true
    ).trim() == 'yes'

    if (vpaInstalled) {
        for (d in deployList) {
            writeFile file: "/tmp/vpa-${d}.yaml", text: """\
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: ${d}-vpa
  namespace: ${namespace}
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ${d}
  updatePolicy:
    updateMode: "Off"
  resourcePolicy:
    containerPolicies:
    - containerName: '*'
      minAllowed:
        cpu: 50m
        memory: 64Mi
      maxAllowed:
        cpu: 4
        memory: 4Gi
"""
            sh "kubectl apply -f /tmp/vpa-${d}.yaml && rm -f /tmp/vpa-${d}.yaml"
            echo "  VPA → ${d}"
        }
    }

    echo "=== K8s enhancements applied: ${namespace} ==="
}
return this
