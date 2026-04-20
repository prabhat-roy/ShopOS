package com.shopos.supplierportalservice.domain;

/**
 * Lifecycle states of a supplier invoice.
 *
 * <p>Allowed transitions:
 * <pre>
 *   DRAFT → SUBMITTED → UNDER_REVIEW → APPROVED → PAID
 *                                     → REJECTED
 *   Any state          → REJECTED (admin override)
 * </pre>
 */
public enum SupplierInvoiceStatus {
    DRAFT,
    SUBMITTED,
    UNDER_REVIEW,
    APPROVED,
    REJECTED,
    PAID
}
