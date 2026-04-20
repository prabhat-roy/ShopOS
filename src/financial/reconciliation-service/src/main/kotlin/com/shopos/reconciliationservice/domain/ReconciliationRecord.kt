package com.shopos.reconciliationservice.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

@Entity
@Table(name = "reconciliation_records")
class ReconciliationRecord {

    @Id
    @Column(name = "id", nullable = false, updatable = false)
    var id: UUID = UUID.randomUUID()

    @Column(name = "internal_payment_id", nullable = false)
    var internalPaymentId: UUID = UUID.randomUUID()

    @Column(name = "external_transaction_id", nullable = false, length = 255)
    var externalTransactionId: String = ""

    @Column(name = "amount", nullable = false, precision = 19, scale = 4)
    var amount: BigDecimal = BigDecimal.ZERO

    @Column(name = "currency", nullable = false, length = 3)
    var currency: String = "USD"

    @Column(name = "internal_amount", nullable = false, precision = 19, scale = 4)
    var internalAmount: BigDecimal = BigDecimal.ZERO

    @Column(name = "external_amount", nullable = false, precision = 19, scale = 4)
    var externalAmount: BigDecimal = BigDecimal.ZERO

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false, length = 20)
    var status: ReconciliationStatus = ReconciliationStatus.UNMATCHED

    /**
     * Computed discrepancy: externalAmount - internalAmount.
     * Zero when MATCHED; non-zero when UNMATCHED/DISPUTED.
     */
    @Column(name = "discrepancy", nullable = false, precision = 19, scale = 4)
    var discrepancy: BigDecimal = BigDecimal.ZERO

    @Column(name = "processor", nullable = false, length = 100)
    var processor: String = ""

    @Column(name = "reconciled_at")
    var reconciledAt: LocalDateTime? = null

    @Column(name = "notes", columnDefinition = "TEXT")
    var notes: String? = null

    @Column(name = "created_at", nullable = false, updatable = false)
    var createdAt: LocalDateTime = LocalDateTime.now()

    @Column(name = "updated_at", nullable = false)
    var updatedAt: LocalDateTime = LocalDateTime.now()

    @PreUpdate
    fun onUpdate() {
        updatedAt = LocalDateTime.now()
    }
}
