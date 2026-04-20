package com.shopos.reconciliationservice.service

import com.shopos.reconciliationservice.domain.ReconciliationRecord
import com.shopos.reconciliationservice.domain.ReconciliationStatus
import com.shopos.reconciliationservice.dto.*
import com.shopos.reconciliationservice.repository.ReconciliationRepository
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

@Service
@Transactional
class ReconciliationService(
    private val reconciliationRepository: ReconciliationRepository
) {

    private val log = LoggerFactory.getLogger(ReconciliationService::class.java)

    // ── Core reconciliation ───────────────────────────────────────────────────

    fun reconcile(request: ReconcileRequest): ReconciliationResponse {
        // Idempotency: skip if already reconciled for this payment
        reconciliationRepository.findByInternalPaymentId(request.internalPaymentId)
            .ifPresent { existing ->
                log.info("Reconciliation already exists for internalPaymentId={}", request.internalPaymentId)
                throw IllegalStateException(
                    "Reconciliation record already exists for internalPaymentId=${request.internalPaymentId}"
                )
            }

        val discrepancy = request.externalAmount.subtract(request.internalAmount)
        val status = if (discrepancy.compareTo(BigDecimal.ZERO) == 0) {
            ReconciliationStatus.MATCHED
        } else {
            ReconciliationStatus.UNMATCHED
        }

        val record = ReconciliationRecord().apply {
            internalPaymentId   = request.internalPaymentId
            externalTransactionId = request.externalTransactionId
            amount              = request.internalAmount       // canonical amount from our system
            currency            = request.currency
            internalAmount      = request.internalAmount
            externalAmount      = request.externalAmount
            this.discrepancy    = discrepancy.abs()
            this.status         = status
            processor           = request.processor
            if (status == ReconciliationStatus.MATCHED) {
                reconciledAt = LocalDateTime.now()
            }
        }

        log.info(
            "Reconciling paymentId={} processor={} status={}",
            request.internalPaymentId, request.processor, status
        )
        return ReconciliationResponse.from(reconciliationRepository.save(record))
    }

    @Transactional(readOnly = true)
    fun getRecord(id: UUID): ReconciliationResponse {
        val record = reconciliationRepository.findById(id)
            .orElseThrow { NoSuchElementException("Reconciliation record not found: $id") }
        return ReconciliationResponse.from(record)
    }

    @Transactional(readOnly = true)
    fun listRecords(
        status: ReconciliationStatus?,
        processor: String?
    ): List<ReconciliationResponse> {
        val records = when {
            status != null && processor != null ->
                // Use a date range spanning all records when no date filter required
                reconciliationRepository.findByProcessorAndStatusAndCreatedAtBetween(
                    processor, status,
                    LocalDateTime.of(2000, 1, 1, 0, 0),
                    LocalDateTime.now().plusDays(1)
                )
            status != null ->
                reconciliationRepository.findByStatus(status)
            processor != null ->
                reconciliationRepository.findByProcessorAndCreatedAtBetween(
                    processor,
                    LocalDateTime.of(2000, 1, 1, 0, 0),
                    LocalDateTime.now().plusDays(1)
                )
            else ->
                reconciliationRepository.findAll()
        }
        return records.map { ReconciliationResponse.from(it) }
    }

    fun disputeRecord(id: UUID, request: DisputeRequest): ReconciliationResponse {
        val record = reconciliationRepository.findById(id)
            .orElseThrow { NoSuchElementException("Reconciliation record not found: $id") }

        if (record.status != ReconciliationStatus.UNMATCHED) {
            throw IllegalStateException(
                "Only UNMATCHED records can be disputed. Current status: ${record.status}"
            )
        }

        record.status = ReconciliationStatus.DISPUTED
        record.notes = request.reason
        record.updatedAt = LocalDateTime.now()

        log.info("Record id={} moved to DISPUTED", id)
        return ReconciliationResponse.from(reconciliationRepository.save(record))
    }

    fun resolveDispute(id: UUID): ReconciliationResponse {
        val record = reconciliationRepository.findById(id)
            .orElseThrow { NoSuchElementException("Reconciliation record not found: $id") }

        if (record.status != ReconciliationStatus.DISPUTED) {
            throw IllegalStateException(
                "Only DISPUTED records can be resolved. Current status: ${record.status}"
            )
        }

        record.status = ReconciliationStatus.RESOLVED
        record.reconciledAt = LocalDateTime.now()
        record.updatedAt = LocalDateTime.now()

        log.info("Record id={} moved to RESOLVED", id)
        return ReconciliationResponse.from(reconciliationRepository.save(record))
    }

    @Transactional(readOnly = true)
    fun getSummary(
        processor: String?,
        startDate: LocalDateTime,
        endDate: LocalDateTime
    ): ReconciliationSummary {
        val records = if (processor != null) {
            reconciliationRepository.findByProcessorAndCreatedAtBetween(processor, startDate, endDate)
        } else {
            reconciliationRepository.findByCreatedAtBetween(startDate, endDate)
        }

        val byStatus = records
            .groupBy { it.status }
            .map { (status, group) ->
                StatusCount(
                    status = status,
                    count = group.size.toLong(),
                    totalDiscrepancy = group.fold(BigDecimal.ZERO) { acc, r -> acc + r.discrepancy }
                )
            }
            .sortedBy { it.status.name }

        val totalDiscrepancy = records.fold(BigDecimal.ZERO) { acc, r -> acc + r.discrepancy }

        return ReconciliationSummary(
            processor = processor,
            totalRecords = records.size.toLong(),
            totalDiscrepancy = totalDiscrepancy,
            byStatus = byStatus
        )
    }
}
