package com.shopos.contractservice.service

import com.shopos.contractservice.domain.Contract
import com.shopos.contractservice.domain.ContractStatus
import com.shopos.contractservice.domain.ContractType
import com.shopos.contractservice.dto.CreateContractRequest
import com.shopos.contractservice.exception.NotFoundException
import com.shopos.contractservice.repository.ContractRepository
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.InjectMocks
import org.mockito.Mock
import org.mockito.junit.jupiter.MockitoExtension
import org.mockito.kotlin.any
import org.mockito.kotlin.argumentCaptor
import org.mockito.kotlin.never
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import java.math.BigDecimal
import java.time.LocalDate
import java.time.LocalDateTime
import java.util.Optional
import java.util.UUID

@ExtendWith(MockitoExtension::class)
class ContractServiceTest {

    @Mock
    private lateinit var contractRepository: ContractRepository

    @InjectMocks
    private lateinit var contractService: ContractService

    // ── helpers ───────────────────────────────────────────────────────────────

    private fun buildContract(
        id: UUID = UUID.randomUUID(),
        status: ContractStatus = ContractStatus.DRAFT,
        startDate: LocalDate = LocalDate.now().minusDays(1),
        endDate: LocalDate = LocalDate.now().plusDays(365)
    ) = Contract(
        id = id,
        orgId = UUID.randomUUID(),
        title = "Master Supply Agreement",
        type = ContractType.SUPPLY_AGREEMENT,
        status = status,
        startDate = startDate,
        endDate = endDate,
        currency = "USD",
        createdBy = "user@acme.com",
        createdAt = LocalDateTime.now(),
        updatedAt = LocalDateTime.now()
    )

    // ── Test 1: createContract saves with DRAFT status ────────────────────────

    @Test
    @DisplayName("createContract: should persist contract with DRAFT status")
    fun `createContract saves contract with DRAFT status`() {
        val orgId = UUID.randomUUID()
        val request = CreateContractRequest(
            title = "NDA Agreement",
            orgId = orgId,
            type = ContractType.NDA,
            startDate = LocalDate.now(),
            endDate = LocalDate.now().plusDays(365),
            createdBy = "user@acme.com",
            value = BigDecimal("50000.00")
        )
        val saved = buildContract(status = ContractStatus.DRAFT)
        whenever(contractRepository.save(any())).thenReturn(saved)

        val captor = argumentCaptor<Contract>()
        whenever(contractRepository.save(captor.capture())).thenReturn(saved)

        val response = contractService.createContract(request)

        assertThat(captor.firstValue.status).isEqualTo(ContractStatus.DRAFT)
        assertThat(captor.firstValue.orgId).isEqualTo(orgId)
        assertThat(response).isNotNull()
    }

    // ── Test 2: getContract found ─────────────────────────────────────────────

    @Test
    @DisplayName("getContract: should return ContractResponse when found")
    fun `getContract returns response when contract exists`() {
        val id = UUID.randomUUID()
        val contract = buildContract(id = id, status = ContractStatus.ACTIVE)
        whenever(contractRepository.findById(id)).thenReturn(Optional.of(contract))

        val response = contractService.getContract(id)

        assertThat(response.id).isEqualTo(id)
        assertThat(response.status).isEqualTo(ContractStatus.ACTIVE)
    }

    // ── Test 3: getContract not found ─────────────────────────────────────────

    @Test
    @DisplayName("getContract: should throw NotFoundException when contract does not exist")
    fun `getContract throws NotFoundException when not found`() {
        val id = UUID.randomUUID()
        whenever(contractRepository.findById(id)).thenReturn(Optional.empty())

        assertThatThrownBy { contractService.getContract(id) }
            .isInstanceOf(NotFoundException::class.java)
            .hasMessageContaining(id.toString())
    }

    // ── Test 4: submitForReview happy path ────────────────────────────────────

    @Test
    @DisplayName("submitForReview: DRAFT contract transitions to UNDER_REVIEW")
    fun `submitForReview transitions DRAFT to UNDER_REVIEW`() {
        val id = UUID.randomUUID()
        val contract = buildContract(id = id, status = ContractStatus.DRAFT)
        whenever(contractRepository.findById(id)).thenReturn(Optional.of(contract))
        whenever(contractRepository.save(contract)).thenReturn(contract)

        contractService.submitForReview(id)

        assertThat(contract.status).isEqualTo(ContractStatus.UNDER_REVIEW)
        verify(contractRepository).save(contract)
    }

    // ── Test 5: submitForReview wrong status ──────────────────────────────────

    @Test
    @DisplayName("submitForReview: should throw when contract is not DRAFT")
    fun `submitForReview throws when status is not DRAFT`() {
        val id = UUID.randomUUID()
        val contract = buildContract(id = id, status = ContractStatus.ACTIVE)
        whenever(contractRepository.findById(id)).thenReturn(Optional.of(contract))

        assertThatThrownBy { contractService.submitForReview(id) }
            .isInstanceOf(IllegalStateException::class.java)
            .hasMessageContaining("DRAFT")

        verify(contractRepository, never()).save(any())
    }

    // ── Test 6: approve transitions UNDER_REVIEW to APPROVED ─────────────────

    @Test
    @DisplayName("approve: UNDER_REVIEW contract transitions to APPROVED")
    fun `approve transitions UNDER_REVIEW to APPROVED`() {
        val id = UUID.randomUUID()
        val contract = buildContract(id = id, status = ContractStatus.UNDER_REVIEW)
        whenever(contractRepository.findById(id)).thenReturn(Optional.of(contract))
        whenever(contractRepository.save(contract)).thenReturn(contract)

        contractService.approve(id)

        assertThat(contract.status).isEqualTo(ContractStatus.APPROVED)
    }

    // ── Test 7: activate APPROVED contract with valid startDate ───────────────

    @Test
    @DisplayName("activate: APPROVED contract with startDate <= today transitions to ACTIVE")
    fun `activate transitions APPROVED to ACTIVE when startDate is in the past`() {
        val id = UUID.randomUUID()
        val contract = buildContract(
            id = id,
            status = ContractStatus.APPROVED,
            startDate = LocalDate.now().minusDays(1)
        )
        whenever(contractRepository.findById(id)).thenReturn(Optional.of(contract))
        whenever(contractRepository.save(contract)).thenReturn(contract)

        contractService.activate(id)

        assertThat(contract.status).isEqualTo(ContractStatus.ACTIVE)
    }

    // ── Test 8: signBuyer sets signedAt when both parties signed ──────────────

    @Test
    @DisplayName("signBuyer: should set signedAt when vendor has already signed")
    fun `signBuyer sets signedAt when vendor has already signed`() {
        val id = UUID.randomUUID()
        val contract = buildContract(id = id, status = ContractStatus.ACTIVE).also {
            it.signedByVendor = true
        }
        whenever(contractRepository.findById(id)).thenReturn(Optional.of(contract))
        whenever(contractRepository.save(contract)).thenReturn(contract)

        contractService.signBuyer(id)

        assertThat(contract.signedByBuyer).isTrue()
        assertThat(contract.signedAt).isNotNull()
    }

    // ── Test 9: terminate sets reason and TERMINATED status ───────────────────

    @Test
    @DisplayName("terminate: should set TERMINATED status and persist reason")
    fun `terminate sets TERMINATED status and stores reason`() {
        val id = UUID.randomUUID()
        val contract = buildContract(id = id, status = ContractStatus.ACTIVE)
        whenever(contractRepository.findById(id)).thenReturn(Optional.of(contract))
        whenever(contractRepository.save(contract)).thenReturn(contract)

        contractService.terminate(id, "Vendor breached SLA terms")

        assertThat(contract.status).isEqualTo(ContractStatus.TERMINATED)
        assertThat(contract.terminationReason).isEqualTo("Vendor breached SLA terms")
    }

    // ── Test 10: detectExpired marks ACTIVE past-endDate contracts EXPIRED ────

    @Test
    @DisplayName("detectExpired: should set EXPIRED on all ACTIVE contracts past endDate and return count")
    fun `detectExpired marks expired contracts and returns count`() {
        val orgId = UUID.randomUUID()
        val expired1 = buildContract(
            status = ContractStatus.ACTIVE,
            endDate = LocalDate.now().minusDays(5)
        )
        val expired2 = buildContract(
            status = ContractStatus.ACTIVE,
            endDate = LocalDate.now().minusDays(1)
        )
        whenever(
            contractRepository.findByEndDateBeforeAndStatus(LocalDate.now(), ContractStatus.ACTIVE)
        ).thenReturn(listOf(expired1, expired2))
        whenever(contractRepository.saveAll(any<List<Contract>>())).thenReturn(listOf(expired1, expired2))

        val count = contractService.detectExpired()

        assertThat(count).isEqualTo(2)
        assertThat(expired1.status).isEqualTo(ContractStatus.EXPIRED)
        assertThat(expired2.status).isEqualTo(ContractStatus.EXPIRED)
    }
}
