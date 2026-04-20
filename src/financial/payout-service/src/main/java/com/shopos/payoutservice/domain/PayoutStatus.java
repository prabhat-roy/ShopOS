package com.shopos.payoutservice.domain;

/**
 * Lifecycle states of a vendor Payout.
 *
 * <pre>
 * PENDING → PROCESSING → COMPLETED
 *                      ↘ FAILED → PENDING (retry)
 * PENDING → CANCELLED
 * </pre>
 */
public enum PayoutStatus {
    /** Payout has been created and is awaiting processing. */
    PENDING,

    /** Payout is actively being processed by the payment rail. */
    PROCESSING,

    /** Funds have been successfully transferred to the vendor. */
    COMPLETED,

    /** Processing failed; the vendor will need to be retried or contacted. */
    FAILED,

    /** Payout was cancelled before processing began. */
    CANCELLED
}
