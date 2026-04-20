def call() {
    def failures  = []
    def warnings  = []

    // ── helper: kubectl rollout status with short timeout ─────────────────────
    def checkRollout = { String kind, String name, String namespace ->
        def rc = sh(
            script: "kubectl rollout status ${kind}/${name} -n ${namespace} --timeout=15s 2>&1",
            returnStatus: true
        )
        if (rc != 0) {
            failures << "${kind}/${name} not ready in namespace '${namespace}'"
        } else {
            echo "  OK  ${kind}/${name} (${namespace})"
        }
    }

    // ── helper: kubectl get with existence check ───────────────────────────────
    def checkExists = { String resource, String name, String namespace ->
        def rc = sh(
            script: "kubectl get ${resource} ${name} -n ${namespace} 2>/dev/null",
            returnStatus: true
        )
        if (rc != 0) {
            failures << "${resource}/${name} not found in namespace '${namespace}'"
        } else {
            echo "  OK  ${resource}/${name} (${namespace})"
        }
    }

    echo "════════════════════════════════════════════════════════════"
    echo " PRE-FLIGHT CHECKS — ${env.BUILD_DOMAIN} → ${env.BUILD_ENV}"
    echo "════════════════════════════════════════════════════════════"

    // ── 1. Target namespace ────────────────────────────────────────────────────
    echo "\n── 1. Domain namespace ──"
    def nsRc = sh(
        script: "kubectl get namespace ${env.BUILD_DOMAIN} 2>/dev/null",
        returnStatus: true
    )
    if (nsRc != 0) {
        failures << "Namespace '${env.BUILD_DOMAIN}' does not exist — run cluster-bootstrap first"
    } else {
        echo "  OK  namespace '${env.BUILD_DOMAIN}' exists"

        def istioLabel = sh(
            script: "kubectl get namespace ${env.BUILD_DOMAIN} -o jsonpath='{.metadata.labels.istio-injection}' 2>/dev/null",
            returnStdout: true
        ).trim()
        if (istioLabel != "enabled") {
            warnings << "Namespace '${env.BUILD_DOMAIN}' missing label istio-injection=enabled — mTLS sidecars will not be injected"
        } else {
            echo "  OK  istio-injection=enabled on namespace '${env.BUILD_DOMAIN}'"
        }
    }

    // ── 2. CNI — Cilium ────────────────────────────────────────────────────────
    echo "\n── 2. CNI (Cilium) ──"
    def ciliumNs = sh(script: "kubectl get daemonset -A --no-headers 2>/dev/null | grep cilium | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!ciliumNs) {
        failures << "Cilium DaemonSet not found — cluster has no CNI"
    } else {
        checkRollout("daemonset", "cilium", ciliumNs)
    }

    // ── 3. Ingress — Traefik ───────────────────────────────────────────────────
    echo "\n── 3. Ingress (Traefik) ──"
    def traefikNs = sh(script: "kubectl get deployment -A --no-headers 2>/dev/null | grep traefik | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!traefikNs) {
        failures << "Traefik Deployment not found — no ingress controller running"
    } else {
        checkRollout("deployment", "traefik", traefikNs)
    }

    // ── 4. Service mesh — Istio ────────────────────────────────────────────────
    echo "\n── 4. Service Mesh (Istio) ──"
    def istiodRc = sh(script: "kubectl get deployment istiod -n istio-system 2>/dev/null", returnStatus: true)
    if (istiodRc != 0) {
        failures << "Istio control plane (istiod) not found in namespace 'istio-system'"
    } else {
        checkRollout("deployment", "istiod", "istio-system")
    }

    // ── 5. TLS — cert-manager ──────────────────────────────────────────────────
    echo "\n── 5. TLS (cert-manager) ──"
    def certmgrNs = sh(script: "kubectl get deployment -A --no-headers 2>/dev/null | grep 'cert-manager ' | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!certmgrNs) {
        failures << "cert-manager Deployment not found — TLS certificate issuance unavailable"
    } else {
        checkRollout("deployment", "cert-manager", certmgrNs)
    }

    // ── 6. Secrets — Vault ────────────────────────────────────────────────────
    echo "\n── 6. Secrets (Vault) ──"
    def vaultNs = sh(script: "kubectl get statefulset -A --no-headers 2>/dev/null | grep vault | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!vaultNs) {
        failures << "Vault StatefulSet not found — secrets management unavailable"
    } else {
        checkRollout("statefulset", "vault", vaultNs)

        // Verify Vault is unsealed
        def sealedRc = sh(
            script: """
                kubectl exec vault-0 -n ${vaultNs} -- vault status -format=json 2>/dev/null \
                | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(1 if d.get('sealed', True) else 0)" 2>/dev/null
            """,
            returnStatus: true
        )
        if (sealedRc != 0) {
            failures << "Vault is sealed — run: kubectl exec vault-0 -n ${vaultNs} -- vault operator unseal"
        } else {
            echo "  OK  Vault is unsealed"
        }
    }

    // ── 7. Secrets sync — External Secrets Operator ───────────────────────────
    echo "\n── 7. Secrets Sync (External Secrets Operator) ──"
    def esoNs = sh(script: "kubectl get deployment -A --no-headers 2>/dev/null | grep external-secrets | grep -v webhook | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!esoNs) {
        failures << "External Secrets Operator not found — K8s Secrets cannot be synced from Vault"
    } else {
        checkRollout("deployment", "external-secrets", esoNs)
    }

    // ── 8. IAM — Keycloak ─────────────────────────────────────────────────────
    echo "\n── 8. IAM (Keycloak) ──"
    def keycloakNs = sh(script: "kubectl get statefulset -A --no-headers 2>/dev/null | grep keycloak | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!keycloakNs) {
        warnings << "Keycloak StatefulSet not found — OIDC/SSO unavailable (non-blocking for initial deploy)"
    } else {
        def kcRc = sh(script: "kubectl rollout status statefulset/keycloak -n ${keycloakNs} --timeout=15s 2>&1", returnStatus: true)
        if (kcRc != 0) { warnings << "Keycloak not fully ready (non-blocking)" }
        else            { echo "  OK  Keycloak (${keycloakNs})" }
    }

    // ── 9. Admission policy — Kyverno ─────────────────────────────────────────
    echo "\n── 9. Admission Policy (Kyverno) ──"
    def kyvernoNs = sh(script: "kubectl get deployment -A --no-headers 2>/dev/null | grep kyverno | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!kyvernoNs) {
        warnings << "Kyverno not found — cluster policies will not be enforced"
    } else {
        checkRollout("deployment", "kyverno", kyvernoNs)
    }

    // ── 10. Messaging — Kafka ─────────────────────────────────────────────────
    echo "\n── 10. Messaging (Kafka) ──"
    def kafkaNs = sh(script: "kubectl get statefulset -A --no-headers 2>/dev/null | grep kafka | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!kafkaNs) {
        failures << "Kafka StatefulSet not found — services that consume Kafka topics will fail to start"
    } else {
        checkRollout("statefulset", "kafka", kafkaNs)
    }

    // ── 11. Messaging — Schema Registry ──────────────────────────────────────
    echo "\n── 11. Messaging (Schema Registry) ──"
    def srNs = sh(script: "kubectl get deployment -A --no-headers 2>/dev/null | grep schema-registry | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!srNs) {
        failures << "Schema Registry Deployment not found — Avro serialisation will fail"
    } else {
        checkRollout("deployment", "schema-registry", srNs)
    }

    // ── 12. Observability — Prometheus ────────────────────────────────────────
    echo "\n── 12. Observability (Prometheus) ──"
    def promNs = sh(script: "kubectl get deployment -A --no-headers 2>/dev/null | grep prometheus-server | awk '{print \$1}' | head -1", returnStdout: true).trim()
    if (!promNs) {
        warnings << "Prometheus not found — metrics will not be collected (non-blocking)"
    } else {
        def promRc = sh(script: "kubectl rollout status deployment/prometheus-server -n ${promNs} --timeout=15s 2>&1", returnStatus: true)
        if (promRc != 0) { warnings << "Prometheus not ready (non-blocking)" }
        else             { echo "  OK  Prometheus (${promNs})" }
    }

    // ── 13. Container registry — Harbor reachable ─────────────────────────────
    echo "\n── 13. Container Registry (Harbor) ──"
    if (env.HARBOR_URL) {
        def harborRc = sh(
            script: "curl -sf --max-time 10 https://${env.HARBOR_URL}/api/v2.0/ping 2>/dev/null || curl -sf --max-time 10 http://${env.HARBOR_URL}/api/v2.0/ping 2>/dev/null",
            returnStatus: true
        )
        if (harborRc != 0) {
            failures << "Harbor registry at '${env.HARBOR_URL}' is unreachable — images cannot be pushed or pulled"
        } else {
            echo "  OK  Harbor registry (${env.HARBOR_URL})"
        }
    } else {
        warnings << "HARBOR_URL not set in infra.env — registry check skipped"
    }

    // ── Summary ───────────────────────────────────────────────────────────────
    echo "\n════════════════════════════════════════════════════════════"

    if (warnings) {
        echo " WARNINGS (${warnings.size()}) — deployment will proceed:"
        warnings.each { echo "  WARN  ${it}" }
    }

    if (failures) {
        echo "\n FAILURES (${failures.size()}) — deployment blocked:"
        failures.each { echo "  FAIL  ${it}" }
        echo """
════════════════════════════════════════════════════════════
  One or more required components are not ready.
  Run the cluster-bootstrap pipeline first, then re-trigger
  this deploy pipeline.

  To skip a component group use SKIP_* flags on cluster-bootstrap.
════════════════════════════════════════════════════════════"""
        error("Pre-flight checks failed: ${failures.size()} blocking issue(s). See FAIL lines above.")
    }

    echo " All required components are ready. Proceeding with deployment."
    echo "════════════════════════════════════════════════════════════\n"
}

return this
