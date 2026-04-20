package com.shopos.invoiceservice.domain;

/**
 * Lifecycle states of an Invoice.
 *
 * <pre>
 * DRAFT → ISSUED → SENT → PAID
 *                ↘ OVERDUE
 * Any → CANCELLED
 * Any → VOID
 * </pre>
 */
public enum InvoiceStatus {
    /** Invoice has been created but not yet finalised or sent to the customer. */
    DRAFT,

    /** Invoice has been finalised and is ready to be sent. */
    ISSUED,

    /** Invoice has been dispatched to the customer. */
    SENT,

    /** Payment has been received in full. */
    PAID,

    /** Due date has passed without full payment. */
    OVERDUE,

    /** Invoice was cancelled before payment. */
    CANCELLED,

    /** Invoice has been voided (legally nullified after issuance). */
    VOID
}
