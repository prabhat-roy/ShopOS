package com.shopos.contractservice.controller

import com.shopos.contractservice.domain.ContractStatus
import com.shopos.contractservice.dto.ContractResponse
import com.shopos.contractservice.dto.CreateContractRequest
import com.shopos.contractservice.dto.TerminateContractRequest
import com.shopos.contractservice.service.ContractService
import jakarta.validation.Valid
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import java.util.UUID

@RestController
class ContractController(
    private val contractService: ContractService
) {

    @GetMapping("/healthz")
    fun health(): ResponseEntity<Map<String, String>> =
        ResponseEntity.ok(mapOf("status" to "ok"))

    // ─── Contract CRUD ────────────────────────────────────────────────────────

    @PostMapping("/contracts")
    fun createContract(@Valid @RequestBody request: CreateContractRequest): ResponseEntity<ContractResponse> {
        val response = contractService.createContract(request)
        return ResponseEntity.status(HttpStatus.CREATED).body(response)
    }

    @GetMapping("/contracts/{id}")
    fun getContract(@PathVariable id: UUID): ResponseEntity<ContractResponse> =
        ResponseEntity.ok(contractService.getContract(id))

    @GetMapping("/contracts")
    fun listContracts(
        @RequestParam(required = false) orgId: UUID?,
        @RequestParam(required = false) status: ContractStatus?
    ): ResponseEntity<List<ContractResponse>> =
        ResponseEntity.ok(contractService.listContracts(orgId, status))

    // ─── Lifecycle transitions ────────────────────────────────────────────────

    @PatchMapping("/contracts/{id}/submit")
    fun submitForReview(@PathVariable id: UUID): ResponseEntity<Void> {
        contractService.submitForReview(id)
        return ResponseEntity.noContent().build()
    }

    @PatchMapping("/contracts/{id}/approve")
    fun approve(@PathVariable id: UUID): ResponseEntity<Void> {
        contractService.approve(id)
        return ResponseEntity.noContent().build()
    }

    @PatchMapping("/contracts/{id}/activate")
    fun activate(@PathVariable id: UUID): ResponseEntity<Void> {
        contractService.activate(id)
        return ResponseEntity.noContent().build()
    }

    // ─── Signing ──────────────────────────────────────────────────────────────

    @PostMapping("/contracts/{id}/sign-buyer")
    fun signBuyer(@PathVariable id: UUID): ResponseEntity<Void> {
        contractService.signBuyer(id)
        return ResponseEntity.noContent().build()
    }

    @PostMapping("/contracts/{id}/sign-vendor")
    fun signVendor(@PathVariable id: UUID): ResponseEntity<Void> {
        contractService.signVendor(id)
        return ResponseEntity.noContent().build()
    }

    // ─── Termination / Cancellation ───────────────────────────────────────────

    @PostMapping("/contracts/{id}/terminate")
    fun terminate(
        @PathVariable id: UUID,
        @Valid @RequestBody request: TerminateContractRequest
    ): ResponseEntity<Void> {
        contractService.terminate(id, request.reason)
        return ResponseEntity.noContent().build()
    }

    @DeleteMapping("/contracts/{id}")
    fun cancel(@PathVariable id: UUID): ResponseEntity<Void> {
        contractService.cancel(id)
        return ResponseEntity.noContent().build()
    }

    // ─── Batch operations ─────────────────────────────────────────────────────

    @PostMapping("/contracts/detect-expired")
    fun detectExpired(): ResponseEntity<Map<String, Int>> {
        val count = contractService.detectExpired()
        return ResponseEntity.ok(mapOf("expiredCount" to count))
    }
}
