package com.shopos.contractservice.service

import com.shopos.contractservice.domain.Contract
import com.shopos.contractservice.domain.ContractStatus
import com.shopos.contractservice.dto.ContractResponse
import com.shopos.contractservice.dto.CreateContractRequest
import com.shopos.contractservice.exception.NotFoundException
import com.shopos.contractservice.repository.ContractRepository
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.time.LocalDate
import java.time.LocalDateTime
import java.util.UUID

@Service
@Transactional(readOnly = true)
class ContractService(
    private val contractRepository: ContractRepository
) {

    @Transactional
    fun createContract(request: CreateContractRequest): ContractResponse {
        val contract = Contract(
            orgId = request.orgId,
            vendorId = request.vendorId,
            title = request.title,
            type = request.type,
            status = ContractStatus.DRAFT,
            description = request.description,
            terms = request.terms,
            value = request.value,
            currency = request.currency,
            startDate = request.startDate,
            endDate = request.endDate,
            autoRenew = request.autoRenew,
            createdBy = request.createdBy
        )
        return ContractResponse.from(contractRepository.save(contract))
    }

    fun getContract(id: UUID): ContractResponse {
        val contract = contractRepository.findById(id)
            .orElseThrow { NotFoundException("Contract", id) }
        return ContractResponse.from(contract)
    }

    fun listContracts(orgId: UUID?, status: ContractStatus?): List<ContractResponse> {
        val contracts = when {
            orgId != null && status != null -> contractRepository.findByOrgIdAndStatus(orgId, status)
            orgId != null -> contractRepository.findByOrgId(orgId)
            status != null -> contractRepository.findByStatus(status)
            else -> contractRepository.findAll()
        }
        return contracts.map { ContractResponse.from(it) }
    }

    @Transactional
    fun submitForReview(id: UUID): ContractResponse {
        val contract = findOrThrow(id)
        check(contract.status == ContractStatus.DRAFT) {
            "Contract must be in DRAFT status to submit for review, current status: ${contract.status}"
        }
        contract.status = ContractStatus.UNDER_REVIEW
        return ContractResponse.from(contractRepository.save(contract))
    }

    @Transactional
    fun approve(id: UUID): ContractResponse {
        val contract = findOrThrow(id)
        check(contract.status == ContractStatus.UNDER_REVIEW) {
            "Contract must be in UNDER_REVIEW status to approve, current status: ${contract.status}"
        }
        contract.status = ContractStatus.APPROVED
        return ContractResponse.from(contractRepository.save(contract))
    }

    @Transactional
    fun activate(id: UUID): ContractResponse {
        val contract = findOrThrow(id)
        check(contract.status == ContractStatus.APPROVED) {
            "Contract must be in APPROVED status to activate, current status: ${contract.status}"
        }
        check(!contract.startDate.isAfter(LocalDate.now())) {
            "Contract start date ${contract.startDate} has not been reached yet"
        }
        contract.status = ContractStatus.ACTIVE
        return ContractResponse.from(contractRepository.save(contract))
    }

    @Transactional
    fun signBuyer(id: UUID): ContractResponse {
        val contract = findOrThrow(id)
        contract.signedByBuyer = true
        if (contract.signedByVendor) {
            contract.signedAt = LocalDateTime.now()
        }
        return ContractResponse.from(contractRepository.save(contract))
    }

    @Transactional
    fun signVendor(id: UUID): ContractResponse {
        val contract = findOrThrow(id)
        contract.signedByVendor = true
        if (contract.signedByBuyer) {
            contract.signedAt = LocalDateTime.now()
        }
        return ContractResponse.from(contractRepository.save(contract))
    }

    @Transactional
    fun terminate(id: UUID, reason: String): ContractResponse {
        val contract = findOrThrow(id)
        check(contract.status == ContractStatus.ACTIVE || contract.status == ContractStatus.APPROVED) {
            "Only ACTIVE or APPROVED contracts can be terminated, current status: ${contract.status}"
        }
        contract.status = ContractStatus.TERMINATED
        contract.terminationReason = reason
        return ContractResponse.from(contractRepository.save(contract))
    }

    @Transactional
    fun cancel(id: UUID): ContractResponse {
        val contract = findOrThrow(id)
        check(
            contract.status == ContractStatus.DRAFT ||
            contract.status == ContractStatus.UNDER_REVIEW ||
            contract.status == ContractStatus.APPROVED
        ) {
            "Cannot cancel a contract in status: ${contract.status}"
        }
        contract.status = ContractStatus.CANCELLED
        return ContractResponse.from(contractRepository.save(contract))
    }

    @Transactional
    fun detectExpired(): Int {
        val today = LocalDate.now()
        val expiredContracts = contractRepository.findByEndDateBeforeAndStatus(today, ContractStatus.ACTIVE)
        expiredContracts.forEach { it.status = ContractStatus.EXPIRED }
        contractRepository.saveAll(expiredContracts)
        return expiredContracts.size
    }

    // ── Private helpers ───────────────────────────────────────────────────────

    private fun findOrThrow(id: UUID): Contract =
        contractRepository.findById(id).orElseThrow { NotFoundException("Contract", id) }
}
