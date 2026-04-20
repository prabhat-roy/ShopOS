package com.shopos.erpintegrationservice.domain;

/**
 * Lifecycle states for an ERP synchronisation operation.
 */
public enum SyncStatus {
    /** Sync has been accepted but not yet started. */
    PENDING,

    /** Sync is actively processing records. */
    IN_PROGRESS,

    /** All records were synchronised without error. */
    SUCCESS,

    /** Sync aborted due to a fatal error; no records committed. */
    FAILED,

    /** Some records succeeded, others failed; partial data in ERP. */
    PARTIAL
}
