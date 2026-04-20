package com.shopos.accountingservice.service

import com.shopos.accountingservice.domain.Account
import com.shopos.accountingservice.domain.AccountType
import com.shopos.accountingservice.domain.JournalEntry
import com.shopos.accountingservice.domain.JournalLine
import com.shopos.accountingservice.dto.*
import com.shopos.accountingservice.repository.AccountRepository
import com.shopos.accountingservice.repository.JournalEntryRepository
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

@Service
@Transactional
class AccountingService(
    private val accountRepository: AccountRepository,
    private val journalEntryRepository: JournalEntryRepository
) {

    // ── Account operations ────────────────────────────────────────────────────

    fun createAccount(request: CreateAccountRequest): AccountResponse {
        if (accountRepository.existsByCode(request.code)) {
            throw IllegalArgumentException("Account with code '${request.code}' already exists")
        }
        val account = Account().apply {
            code = request.code
            name = request.name
            type = request.type
            currency = request.currency
        }
        return AccountResponse.from(accountRepository.save(account))
    }

    @Transactional(readOnly = true)
    fun getAccount(id: UUID): AccountResponse {
        val account = accountRepository.findById(id)
            .orElseThrow { NoSuchElementException("Account not found: $id") }
        return AccountResponse.from(account)
    }

    @Transactional(readOnly = true)
    fun listAccounts(type: AccountType?, activeOnly: Boolean): List<AccountResponse> {
        val accounts = when {
            type != null && activeOnly -> accountRepository.findByTypeAndActiveTrue(type)
            type != null               -> accountRepository.findByType(type)
            activeOnly                 -> accountRepository.findByActiveTrue()
            else                       -> accountRepository.findAll()
        }
        return accounts.map { AccountResponse.from(it) }
    }

    fun deactivateAccount(id: UUID): AccountResponse {
        val account = accountRepository.findById(id)
            .orElseThrow { NoSuchElementException("Account not found: $id") }
        account.active = false
        account.updatedAt = LocalDateTime.now()
        return AccountResponse.from(accountRepository.save(account))
    }

    // ── Journal Entry operations ──────────────────────────────────────────────

    fun createJournalEntry(request: CreateJournalEntryRequest): JournalEntryResponse {
        if (journalEntryRepository.existsByReference(request.reference)) {
            throw IllegalArgumentException("Journal entry with reference '${request.reference}' already exists")
        }

        // Validate balanced entry: sum(debits) == sum(credits)
        val totalDebits = request.lines
            .filter { it.type == "debit" }
            .fold(BigDecimal.ZERO) { acc, l -> acc + l.amount }
        val totalCredits = request.lines
            .filter { it.type == "credit" }
            .fold(BigDecimal.ZERO) { acc, l -> acc + l.amount }

        if (totalDebits.compareTo(totalCredits) != 0) {
            throw IllegalArgumentException(
                "Journal entry is not balanced: debits=$totalDebits, credits=$totalCredits"
            )
        }

        // Validate all referenced accounts exist
        val accountIds = request.lines.map { it.accountId }.toSet()
        val accounts = accountRepository.findAllById(accountIds)
        val foundIds = accounts.map { it.id }.toSet()
        val missingIds = accountIds - foundIds
        if (missingIds.isNotEmpty()) {
            throw NoSuchElementException("Accounts not found: $missingIds")
        }
        val accountMap = accounts.associateBy { it.id }

        val entry = JournalEntry().apply {
            reference = request.reference
            description = request.description
            totalAmount = totalDebits   // debits == credits, so either works
            currency = request.currency
        }

        val journalLines = request.lines.map { lineReq ->
            JournalLine().apply {
                journalEntry = entry
                accountId = lineReq.accountId
                type = lineReq.type
                amount = lineReq.amount
            }
        }
        entry.lines.addAll(journalLines)

        val savedEntry = journalEntryRepository.save(entry)

        // Update account balances
        // Debit: increases ASSET and EXPENSE; decreases LIABILITY, EQUITY, REVENUE
        // Credit: increases LIABILITY, EQUITY, REVENUE; decreases ASSET and EXPENSE
        request.lines.forEach { lineReq ->
            val account = accountMap[lineReq.accountId]!!
            val delta = when {
                lineReq.type == "debit" && account.type in setOf(AccountType.ASSET, AccountType.EXPENSE)
                    -> lineReq.amount
                lineReq.type == "debit" && account.type in setOf(AccountType.LIABILITY, AccountType.EQUITY, AccountType.REVENUE)
                    -> lineReq.amount.negate()
                lineReq.type == "credit" && account.type in setOf(AccountType.LIABILITY, AccountType.EQUITY, AccountType.REVENUE)
                    -> lineReq.amount
                lineReq.type == "credit" && account.type in setOf(AccountType.ASSET, AccountType.EXPENSE)
                    -> lineReq.amount.negate()
                else -> BigDecimal.ZERO
            }
            account.balance = account.balance + delta
            account.updatedAt = LocalDateTime.now()
            accountRepository.save(account)
        }

        return JournalEntryResponse.from(savedEntry)
    }

    @Transactional(readOnly = true)
    fun getJournalEntry(id: UUID): JournalEntryResponse {
        val entry = journalEntryRepository.findById(id)
            .orElseThrow { NoSuchElementException("Journal entry not found: $id") }
        return JournalEntryResponse.from(entry)
    }

    @Transactional(readOnly = true)
    fun listJournalEntries(startDate: LocalDateTime, endDate: LocalDateTime): List<JournalEntryResponse> {
        return journalEntryRepository.findByCreatedAtBetween(startDate, endDate)
            .map { JournalEntryResponse.from(it) }
    }

    @Transactional(readOnly = true)
    fun getAccountBalance(accountId: UUID): BigDecimal {
        val account = accountRepository.findById(accountId)
            .orElseThrow { NoSuchElementException("Account not found: $accountId") }
        return account.balance
    }

    @Transactional(readOnly = true)
    fun getLedger(accountId: UUID, startDate: LocalDateTime, endDate: LocalDateTime): List<JournalLineResponse> {
        // Verify account exists
        accountRepository.findById(accountId)
            .orElseThrow { NoSuchElementException("Account not found: $accountId") }

        val entries = journalEntryRepository.findByAccountIdAndDateRange(accountId, startDate, endDate)
        return entries.flatMap { entry ->
            entry.lines
                .filter { it.accountId == accountId }
                .map { line ->
                    JournalLineResponse(
                        id = line.id,
                        entryId = entry.id,
                        accountId = line.accountId,
                        type = line.type,
                        amount = line.amount
                    )
                }
        }
    }
}
