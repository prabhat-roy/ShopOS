package com.shopos.reconciliationservice.dto

import com.shopos.reconciliationservice.domain.ReconciliationRecord
import com.shopos.reconciliationservice.domain.ReconciliationStatus
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

data class ReconciliationResponse(
    val id: UUID,
    val internalPaymentId: UUID,
    val externalTransactionId: String,
    val amount: BigDecimal,
    val currency: String,
    val internalAmount: BigDecimal,
    val externalAmount: BigDecimal,
    val status: ReconciliationStatus,
    val discrepancy: BigDecimal,
    val processor: String,
    val reconciledAt: LocalDateTime?,
    val notes: String?,
    val createdAt: LocalDateTime,
    val updatedAt: LocalDateTime
) {
    companion object {
        fun from(record: ReconciliationRecord): ReconciliationResponse = ReconciliationResponse(
            id = record.id,
            internalPaymentId = record.internalPaymentId,
            externalTransactionId = record.externalTransactionId,
            amount = record.amount,
            currency = record.currency,
            internalAmount = record.internalAmount,
            externalAmount = record.externalAmount,
            status = record.status,
            discrepancy = record.discrepancy,
            processor = record.processor,
            reconciledAt = record.reconciledAt,
            notes = record.notes,
            createdAt = record.createdAt,
            updatedAt = record.updatedAt
        )
    }
}
