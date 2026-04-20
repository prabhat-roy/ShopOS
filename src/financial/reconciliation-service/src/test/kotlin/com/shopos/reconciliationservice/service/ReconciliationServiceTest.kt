package com.shopos.reconciliationservice.service

import com.shopos.reconciliationservice.domain.ReconciliationRecord
import com.shopos.reconciliationservice.domain.ReconciliationStatus
import com.shopos.reconciliationservice.dto.DisputeRequest
import com.shopos.reconciliationservice.dto.ReconcileRequest
import com.shopos.reconciliationservice.repository.ReconciliationRepository
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.*
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.Optional
import java.util.UUID

class ReconciliationServiceTest {

    private val reconciliationRepository: ReconciliationRepository = mock()
    private val service = ReconciliationService(reconciliationRepository)

    // ── helpers ───────────────────────────────────────────────────────────────

    private fun makeRecord(
        status: ReconciliationStatus = ReconciliationStatus.UNMATCHED,
        internalAmount: BigDecimal = BigDecimal("100.00"),
        externalAmount: BigDecimal = BigDecimal("100.00"),
        processor: String = "STRIPE"
    ): ReconciliationRecord = ReconciliationRecord().apply {
        id = UUID.randomUUID()
        internalPaymentId = UUID.randomUUID()
        externalTransactionId = "EXT-${UUID.randomUUID()}"
        amount = internalAmount
        currency = "USD"
        this.internalAmount = internalAmount
        this.externalAmount = externalAmount
        discrepancy = (externalAmount - internalAmount).abs()
        this.status = status
        this.processor = processor
        createdAt = LocalDateTime.now()
        updatedAt = LocalDateTime.now()
    }

    private fun makeRequest(
        internalAmount: BigDecimal = BigDecimal("100.00"),
        externalAmount: BigDecimal = BigDecimal("100.00"),
        processor: String = "STRIPE"
    ): ReconcileRequest = ReconcileRequest(
        internalPaymentId     = UUID.randomUUID(),
        externalTransactionId = "EXT-${UUID.randomUUID()}",
        internalAmount        = internalAmount,
        externalAmount        = externalAmount,
        currency              = "USD",
        processor             = processor
    )

    // ── 1. reconcile — equal amounts → MATCHED ────────────────────────────────

    @Test
    fun `reconcile with equal amounts creates MATCHED record`() {
        val request = makeRequest(BigDecimal("250.00"), BigDecimal("250.00"))
        val saved = makeRecord(ReconciliationStatus.MATCHED, BigDecimal("250.00"), BigDecimal("250.00"))

        whenever(reconciliationRepository.findByInternalPaymentId(request.internalPaymentId))
            .thenReturn(Optional.empty())
        whenever(reconciliationRepository.save(any())).thenReturn(saved)

        val result = service.reconcile(request)

        assertEquals(ReconciliationStatus.MATCHED, result.status)
        verify(reconciliationRepository).save(any())
    }

    // ── 2. reconcile — different amounts → UNMATCHED ──────────────────────────

    @Test
    fun `reconcile with unequal amounts creates UNMATCHED record`() {
        val request = makeRequest(BigDecimal("100.00"), BigDecimal("95.00"))
        val saved = makeRecord(ReconciliationStatus.UNMATCHED, BigDecimal("100.00"), BigDecimal("95.00"))

        whenever(reconciliationRepository.findByInternalPaymentId(request.internalPaymentId))
            .thenReturn(Optional.empty())
        whenever(reconciliationRepository.save(any())).thenReturn(saved)

        val result = service.reconcile(request)

        assertEquals(ReconciliationStatus.UNMATCHED, result.status)
    }

    // ── 3. reconcile — duplicate paymentId throws ─────────────────────────────

    @Test
    fun `reconcile throws when record already exists for paymentId`() {
        val request = makeRequest()
        val existing = makeRecord()

        whenever(reconciliationRepository.findByInternalPaymentId(request.internalPaymentId))
            .thenReturn(Optional.of(existing))

        assertThrows<IllegalStateException> { service.reconcile(request) }
        verify(reconciliationRepository, never()).save(any())
    }

    // ── 4. getRecord — happy path ─────────────────────────────────────────────

    @Test
    fun `getRecord returns response when found`() {
        val record = makeRecord()
        whenever(reconciliationRepository.findById(record.id)).thenReturn(Optional.of(record))

        val result = service.getRecord(record.id)

        assertEquals(record.id, result.id)
        assertEquals(record.processor, result.processor)
    }

    // ── 5. getRecord — not found throws ───────────────────────────────────────

    @Test
    fun `getRecord throws NoSuchElementException when not found`() {
        val id = UUID.randomUUID()
        whenever(reconciliationRepository.findById(id)).thenReturn(Optional.empty())

        assertThrows<NoSuchElementException> { service.getRecord(id) }
    }

    // ── 6. disputeRecord — UNMATCHED → DISPUTED ───────────────────────────────

    @Test
    fun `disputeRecord transitions UNMATCHED record to DISPUTED`() {
        val record = makeRecord(status = ReconciliationStatus.UNMATCHED)
        whenever(reconciliationRepository.findById(record.id)).thenReturn(Optional.of(record))
        whenever(reconciliationRepository.save(any<ReconciliationRecord>()))
            .thenAnswer { it.arguments[0] as ReconciliationRecord }

        val result = service.disputeRecord(record.id, DisputeRequest(reason = "Amount mismatch due to FX rounding"))

        assertEquals(ReconciliationStatus.DISPUTED, result.status)
        assertNotNull(result.notes)
        verify(reconciliationRepository).save(argThat { status == ReconciliationStatus.DISPUTED })
    }

    // ── 7. disputeRecord — non-UNMATCHED throws ───────────────────────────────

    @Test
    fun `disputeRecord throws when record is not UNMATCHED`() {
        val record = makeRecord(status = ReconciliationStatus.MATCHED)
        whenever(reconciliationRepository.findById(record.id)).thenReturn(Optional.of(record))

        assertThrows<IllegalStateException> {
            service.disputeRecord(record.id, DisputeRequest(reason = "This should fail"))
        }
        verify(reconciliationRepository, never()).save(any())
    }

    // ── 8. resolveDispute — DISPUTED → RESOLVED ───────────────────────────────

    @Test
    fun `resolveDispute transitions DISPUTED record to RESOLVED with reconciledAt set`() {
        val record = makeRecord(status = ReconciliationStatus.DISPUTED)
        whenever(reconciliationRepository.findById(record.id)).thenReturn(Optional.of(record))
        whenever(reconciliationRepository.save(any<ReconciliationRecord>()))
            .thenAnswer { it.arguments[0] as ReconciliationRecord }

        val result = service.resolveDispute(record.id)

        assertEquals(ReconciliationStatus.RESOLVED, result.status)
        assertNotNull(result.reconciledAt)
        verify(reconciliationRepository).save(argThat {
            status == ReconciliationStatus.RESOLVED && reconciledAt != null
        })
    }
}
