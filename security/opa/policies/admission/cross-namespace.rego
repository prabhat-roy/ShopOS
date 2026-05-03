package shopos.admission.crossns

import future.keywords.if
import future.keywords.in

# Block creation of a Service of type ExternalName that points to another namespace
# unless the source namespace is in the platform team allowlist.
deny[msg] if {
    input.request.kind.kind == "Service"
    input.request.object.spec.type == "ExternalName"
    en := input.request.object.spec.externalName
    contains(en, ".svc.cluster.local")
    target_ns := split(en, ".")[1]
    src_ns    := input.request.namespace
    target_ns != src_ns
    not src_ns in {"platform","istio-system","monitoring","infra"}
    msg := sprintf("Cross-namespace ExternalName from %q to %q not allowed", [src_ns, target_ns])
}
