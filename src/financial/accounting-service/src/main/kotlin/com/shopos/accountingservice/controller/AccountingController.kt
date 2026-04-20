package com.shopos.accountingservice.controller

import com.shopos.accountingservice.domain.AccountType
import com.shopos.accountingservice.dto.*
import com.shopos.accountingservice.service.AccountingService
import jakarta.validation.Valid
import org.springframework.format.annotation.DateTimeFormat
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

@RestController
class AccountingController(
    private val accountingService: AccountingService
) {

    // ── Health ────────────────────────────────────────────────────────────────

    @GetMapping("/healthz")
    fun health(): ResponseEntity<Map<String, String>> =
        ResponseEntity.ok(mapOf("status" to "ok"))

    // ── Accounts ──────────────────────────────────────────────────────────────

    @PostMapping("/accounts")
    fun createAccount(
        @Valid @RequestBody request: CreateAccountRequest
    ): ResponseEntity<AccountResponse> {
        val response = accountingService.createAccount(request)
        return ResponseEntity.status(HttpStatus.CREATED).body(response)
    }

    @GetMapping("/accounts/{id}")
    fun getAccount(@PathVariable id: UUID): ResponseEntity<AccountResponse> {
        val response = accountingService.getAccount(id)
        return ResponseEntity.ok(response)
    }

    @GetMapping("/accounts")
    fun listAccounts(
        @RequestParam(required = false) type: AccountType?,
        @RequestParam(required = false, defaultValue = "false") activeOnly: Boolean
    ): ResponseEntity<List<AccountResponse>> {
        val response = accountingService.listAccounts(type, activeOnly)
        return ResponseEntity.ok(response)
    }

    @PostMapping("/accounts/{id}/deactivate")
    fun deactivateAccount(@PathVariable id: UUID): ResponseEntity<AccountResponse> {
        val response = accountingService.deactivateAccount(id)
        return ResponseEntity.ok(response)
    }

    @GetMapping("/accounts/{id}/balance")
    fun getAccountBalance(@PathVariable id: UUID): ResponseEntity<Map<String, Any>> {
        val balance = accountingService.getAccountBalance(id)
        return ResponseEntity.ok(mapOf("accountId" to id, "balance" to balance))
    }

    @GetMapping("/accounts/{id}/ledger")
    fun getLedger(
        @PathVariable id: UUID,
        @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) start: LocalDateTime,
        @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) end: LocalDateTime
    ): ResponseEntity<List<JournalLineResponse>> {
        val lines = accountingService.getLedger(id, start, end)
        return ResponseEntity.ok(lines)
    }

    // ── Journal Entries ───────────────────────────────────────────────────────

    @PostMapping("/journal-entries")
    fun createJournalEntry(
        @Valid @RequestBody request: CreateJournalEntryRequest
    ): ResponseEntity<JournalEntryResponse> {
        val response = accountingService.createJournalEntry(request)
        return ResponseEntity.status(HttpStatus.CREATED).body(response)
    }

    @GetMapping("/journal-entries/{id}")
    fun getJournalEntry(@PathVariable id: UUID): ResponseEntity<JournalEntryResponse> {
        val response = accountingService.getJournalEntry(id)
        return ResponseEntity.ok(response)
    }

    @GetMapping("/journal-entries")
    fun listJournalEntries(
        @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) start: LocalDateTime,
        @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) end: LocalDateTime
    ): ResponseEntity<List<JournalEntryResponse>> {
        val response = accountingService.listJournalEntries(start, end)
        return ResponseEntity.ok(response)
    }
}
