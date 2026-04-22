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

    def deployList   = deploys      ? deploys.split()      as List : []
    def statefulList = statefulsets ? statefulsets.split() as List : []
    def allWorkloads = deployList + statefulList

    if (allWorkloads.isEmpty()) {
        echo "  No deployments or statefulsets found in ${namespace} — skipping"
        return
    }

    // Deduplicate Helm-style names: kafka-kafka → kafka, zookeeper-zookeeper → zookeeper
    def shortName = { String w ->
        def m = w =~ /^(.+)-\1$/
        m ? m[0][1] : w
    }

    def kedaInstalled = sh(
        script: 'kubectl get crd scaledobjects.keda.sh >/dev/null 2>&1 && echo yes || echo no',
        returnStdout: true
    ).trim() == 'yes'

    // ── HPA / KEDA ScaledObject for Deployments ───────────────────────────────
    for (d in deployList) {
        def rname = shortName(d)
        if (kedaInstalled) {
            writeFile file: "/tmp/keda-so-${rname}.yaml", text: """\
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: ${rname}-scaledobject
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
            sh "kubectl apply -f /tmp/keda-so-${rname}.yaml && rm -f /tmp/keda-so-${rname}.yaml"
            echo "  KEDA ScaledObject → ${rname} (targets ${d})"
        } else {
            writeFile file: "/tmp/hpa-${rname}.yaml", text: """\
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ${rname}-hpa
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
            sh "kubectl apply -f /tmp/hpa-${rname}.yaml && rm -f /tmp/hpa-${rname}.yaml"
            echo "  HPA → ${rname} (targets ${d})"
        }
    }

    // ── KEDA ScaledObject for StatefulSets (if KEDA available) ───────────────
    for (ss in statefulList) {
        def rname = shortName(ss)
        if (kedaInstalled) {
            writeFile file: "/tmp/keda-so-${rname}.yaml", text: """\
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: ${rname}-scaledobject
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
            sh "kubectl apply -f /tmp/keda-so-${rname}.yaml && rm -f /tmp/keda-so-${rname}.yaml"
            echo "  KEDA ScaledObject (StatefulSet) → ${rname} (targets ${ss})"
        }
    }

    // ── PodDisruptionBudget for all workloads ─────────────────────────────────
    for (w in allWorkloads) {
        def rname = shortName(w)
        def kind = deployList.contains(w) ? 'deployment' : 'statefulset'
        sh """
            SELECTOR=\$(kubectl get ${kind} ${w} -n ${namespace} \
                -o go-template='{{range \$k,\$v := .spec.selector.matchLabels}}{{\$k}}={{\$v}},{{end}}' \
                2>/dev/null | sed 's/,\$//')
            SELECTOR=\${SELECTOR:-app=${w}}
            kubectl create pdb ${rname}-pdb -n ${namespace} \
                --selector="\${SELECTOR}" --min-available=1 \
                --dry-run=client -o yaml | kubectl apply -f -
        """
        echo "  PDB → ${rname} (targets ${w})"
    }

    // ── Vertical Pod Autoscaler (if VPA CRD is available) ────────────────────
    def vpaInstalled = sh(
        script: 'kubectl get crd verticalpodautoscalers.autoscaling.k8s.io >/dev/null 2>&1 && echo yes || echo no',
        returnStdout: true
    ).trim() == 'yes'

    if (vpaInstalled) {
        for (d in deployList) {
            def rname = shortName(d)
            writeFile file: "/tmp/vpa-${rname}.yaml", text: """\
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: ${rname}-vpa
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
            sh "kubectl apply -f /tmp/vpa-${rname}.yaml && rm -f /tmp/vpa-${rname}.yaml"
            echo "  VPA → ${rname} (targets ${d})"
        }
    }

    echo "=== K8s enhancements applied: ${namespace} ==="
}
return this
