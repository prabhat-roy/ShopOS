package com.shopos.reconciliationservice.repository

import com.shopos.reconciliationservice.domain.ReconciliationRecord
import com.shopos.reconciliationservice.domain.ReconciliationStatus
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository
import java.time.LocalDateTime
import java.util.Optional
import java.util.UUID

@Repository
interface ReconciliationRepository : JpaRepository<ReconciliationRecord, UUID> {

    fun findByStatus(status: ReconciliationStatus): List<ReconciliationRecord>

    fun findByInternalPaymentId(internalPaymentId: UUID): Optional<ReconciliationRecord>

    fun findByExternalTransactionId(externalTransactionId: String): Optional<ReconciliationRecord>

    fun findByProcessorAndCreatedAtBetween(
        processor: String,
        start: LocalDateTime,
        end: LocalDateTime
    ): List<ReconciliationRecord>

    fun findByProcessorAndStatusAndCreatedAtBetween(
        processor: String,
        status: ReconciliationStatus,
        start: LocalDateTime,
        end: LocalDateTime
    ): List<ReconciliationRecord>

    fun findByStatusAndCreatedAtBetween(
        status: ReconciliationStatus,
        start: LocalDateTime,
        end: LocalDateTime
    ): List<ReconciliationRecord>

    fun findByCreatedAtBetween(start: LocalDateTime, end: LocalDateTime): List<ReconciliationRecord>

    fun existsByInternalPaymentId(internalPaymentId: UUID): Boolean
}
