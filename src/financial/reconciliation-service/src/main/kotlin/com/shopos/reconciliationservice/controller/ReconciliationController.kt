package com.shopos.reconciliationservice.controller

import com.shopos.reconciliationservice.domain.ReconciliationStatus
import com.shopos.reconciliationservice.dto.*
import com.shopos.reconciliationservice.service.ReconciliationService
import jakarta.validation.Valid
import org.springframework.format.annotation.DateTimeFormat
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import java.time.LocalDateTime
import java.util.UUID

@RestController
class ReconciliationController(
    private val reconciliationService: ReconciliationService
) {

    // ── Health ────────────────────────────────────────────────────────────────

    @GetMapping("/healthz")
    fun health(): ResponseEntity<Map<String, String>> =
        ResponseEntity.ok(mapOf("status" to "ok"))

    // ── Reconciliation ────────────────────────────────────────────────────────

    @PostMapping("/reconcile")
    fun reconcile(
        @Valid @RequestBody request: ReconcileRequest
    ): ResponseEntity<ReconciliationResponse> {
        val response = reconciliationService.reconcile(request)
        return ResponseEntity.status(HttpStatus.CREATED).body(response)
    }

    @GetMapping("/records")
    fun listRecords(
        @RequestParam(required = false) status: ReconciliationStatus?,
        @RequestParam(required = false) processor: String?
    ): ResponseEntity<List<ReconciliationResponse>> {
        val response = reconciliationService.listRecords(status, processor)
        return ResponseEntity.ok(response)
    }

    @GetMapping("/records/{id}")
    fun getRecord(@PathVariable id: UUID): ResponseEntity<ReconciliationResponse> {
        val response = reconciliationService.getRecord(id)
        return ResponseEntity.ok(response)
    }

    @PostMapping("/records/{id}/dispute")
    fun disputeRecord(
        @PathVariable id: UUID,
        @Valid @RequestBody request: DisputeRequest
    ): ResponseEntity<ReconciliationResponse> {
        val response = reconciliationService.disputeRecord(id, request)
        return ResponseEntity.ok(response)
    }

    @PostMapping("/records/{id}/resolve")
    fun resolveDispute(@PathVariable id: UUID): ResponseEntity<ReconciliationResponse> {
        val response = reconciliationService.resolveDispute(id)
        return ResponseEntity.ok(response)
    }

    @GetMapping("/summary")
    fun getSummary(
        @RequestParam(required = false) processor: String?,
        @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) start: LocalDateTime,
        @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) end: LocalDateTime
    ): ResponseEntity<ReconciliationSummary> {
        val summary = reconciliationService.getSummary(processor, start, end)
        return ResponseEntity.ok(summary)
    }
}
