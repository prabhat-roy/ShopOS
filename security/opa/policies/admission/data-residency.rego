package shopos.admission.dataresidency

import future.keywords.if
import future.keywords.in

# Deny if a deployment in the `eu-data` topology references a non-EU storage region.
deny[msg] if {
    input.request.kind.kind == "Deployment"
    input.request.object.metadata.labels["data-residency"] == "eu-only"
    container := input.request.object.spec.template.spec.containers[_]
    env := container.env[_]
    env.name == "S3_REGION"
    not is_eu_region(env.value)
    msg := sprintf("EU-only workload references non-EU region %q", [env.value])
}

is_eu_region(region) if {
    region in {"eu-west-1","eu-west-2","eu-central-1","eu-north-1","eu-south-1","eu-west-3"}
}
