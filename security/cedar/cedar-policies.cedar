// AWS Cedar policies — alternative to OPA for resource-scoped authz (sellers, orgs).
// Loaded by openfga-cedar-bridge in identity domain.

permit (
    principal in Group::"sellers",
    action == Action::"editListing",
    resource
)
when {
    resource.seller_id == principal.seller_id
};

permit (
    principal in Group::"buyers",
    action == Action::"viewOrder",
    resource
)
when {
    resource.user_id == principal.user_id
};

permit (
    principal in Group::"org-admins",
    action in [Action::"approveQuote", Action::"createUser", Action::"editContract"],
    resource
)
when {
    resource.org_id == principal.org_id
};

forbid (
    principal,
    action == Action::"refundPayment",
    resource
)
unless {
    principal in Group::"finance-managers" &&
    resource.amount_cents <= principal.refund_limit_cents
};
