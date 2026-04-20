package com.shopos.accountingservice.service

import com.shopos.accountingservice.domain.Account
import com.shopos.accountingservice.domain.AccountType
import com.shopos.accountingservice.domain.JournalEntry
import com.shopos.accountingservice.domain.JournalLine
import com.shopos.accountingservice.dto.CreateAccountRequest
import com.shopos.accountingservice.dto.CreateJournalEntryRequest
import com.shopos.accountingservice.dto.JournalLineRequest
import com.shopos.accountingservice.repository.AccountRepository
import com.shopos.accountingservice.repository.JournalEntryRepository
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.*
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.Optional
import java.util.UUID

class AccountingServiceTest {

    private val accountRepository: AccountRepository = mock()
    private val journalEntryRepository: JournalEntryRepository = mock()
    private val service = AccountingService(accountRepository, journalEntryRepository)

    // ── helpers ───────────────────────────────────────────────────────────────

    private fun makeAccount(
        type: AccountType = AccountType.ASSET,
        code: String = "CASH",
        balance: BigDecimal = BigDecimal.ZERO
    ): Account = Account().apply {
        this.id = UUID.randomUUID()
        this.code = code
        this.name = "Test Account"
        this.type = type
        this.balance = balance
        this.currency = "USD"
        this.active = true
        this.createdAt = LocalDateTime.now()
        this.updatedAt = LocalDateTime.now()
    }

    private fun makeEntry(vararg lines: JournalLine): JournalEntry = JournalEntry().apply {
        this.id = UUID.randomUUID()
        this.reference = "REF-001"
        this.description = "Test entry"
        this.totalAmount = BigDecimal("100.00")
        this.currency = "USD"
        this.createdAt = LocalDateTime.now()
        this.lines.addAll(lines)
    }

    // ── 1. createAccount — happy path ─────────────────────────────────────────

    @Test
    fun `createAccount returns saved account response`() {
        val request = CreateAccountRequest(code = "CASH", name = "Cash", type = AccountType.ASSET)
        val savedAccount = makeAccount(code = "CASH")

        whenever(accountRepository.existsByCode("CASH")).thenReturn(false)
        whenever(accountRepository.save(any())).thenReturn(savedAccount)

        val result = service.createAccount(request)

        assertEquals("CASH", result.code)
        assertEquals(AccountType.ASSET, result.type)
        verify(accountRepository).save(any())
    }

    // ── 2. createAccount — duplicate code throws ──────────────────────────────

    @Test
    fun `createAccount throws when code already exists`() {
        val request = CreateAccountRequest(code = "CASH", name = "Cash", type = AccountType.ASSET)
        whenever(accountRepository.existsByCode("CASH")).thenReturn(true)

        val ex = assertThrows<IllegalArgumentException> { service.createAccount(request) }
        assertTrue(ex.message!!.contains("CASH"))
        verify(accountRepository, never()).save(any())
    }

    // ── 3. getAccount — happy path ────────────────────────────────────────────

    @Test
    fun `getAccount returns account when found`() {
        val account = makeAccount()
        whenever(accountRepository.findById(account.id)).thenReturn(Optional.of(account))

        val result = service.getAccount(account.id)

        assertEquals(account.id, result.id)
        assertEquals(account.code, result.code)
    }

    // ── 4. getAccount — not found throws ─────────────────────────────────────

    @Test
    fun `getAccount throws NoSuchElementException when not found`() {
        val id = UUID.randomUUID()
        whenever(accountRepository.findById(id)).thenReturn(Optional.empty())

        assertThrows<NoSuchElementException> { service.getAccount(id) }
    }

    // ── 5. listAccounts — filter by type ─────────────────────────────────────

    @Test
    fun `listAccounts filters by type when type provided`() {
        val assetAccount = makeAccount(type = AccountType.ASSET, code = "CASH")
        whenever(accountRepository.findByType(AccountType.ASSET)).thenReturn(listOf(assetAccount))

        val result = service.listAccounts(type = AccountType.ASSET, activeOnly = false)

        assertEquals(1, result.size)
        assertEquals(AccountType.ASSET, result[0].type)
        verify(accountRepository).findByType(AccountType.ASSET)
    }

    // ── 6. createJournalEntry — balanced entry succeeds ───────────────────────

    @Test
    fun `createJournalEntry succeeds with balanced debits and credits`() {
        val cashAccount = makeAccount(type = AccountType.ASSET, code = "CASH")
        val revenueAccount = makeAccount(type = AccountType.REVENUE, code = "SALES-REV")

        val request = CreateJournalEntryRequest(
            reference = "INV-001",
            description = "Sale of goods",
            currency = "USD",
            lines = listOf(
                JournalLineRequest(accountId = cashAccount.id, type = "debit",  amount = BigDecimal("500.00")),
                JournalLineRequest(accountId = revenueAccount.id, type = "credit", amount = BigDecimal("500.00"))
            )
        )

        whenever(journalEntryRepository.existsByReference("INV-001")).thenReturn(false)
        whenever(accountRepository.findAllById(any())).thenReturn(listOf(cashAccount, revenueAccount))

        val savedEntry = makeEntry().apply {
            reference = "INV-001"
            totalAmount = BigDecimal("500.00")
        }
        whenever(journalEntryRepository.save(any())).thenReturn(savedEntry)
        whenever(accountRepository.save(any())).thenReturn(cashAccount)

        val result = service.createJournalEntry(request)

        assertNotNull(result)
        assertEquals("INV-001", result.reference)
        verify(journalEntryRepository).save(any())
        // two account balance updates
        verify(accountRepository, times(2)).save(any())
    }

    // ── 7. createJournalEntry — unbalanced entry throws ───────────────────────

    @Test
    fun `createJournalEntry throws when debits do not equal credits`() {
        val request = CreateJournalEntryRequest(
            reference = "BAD-001",
            description = "Unbalanced entry",
            lines = listOf(
                JournalLineRequest(accountId = UUID.randomUUID(), type = "debit",  amount = BigDecimal("300.00")),
                JournalLineRequest(accountId = UUID.randomUUID(), type = "credit", amount = BigDecimal("200.00"))
            )
        )
        whenever(journalEntryRepository.existsByReference("BAD-001")).thenReturn(false)

        val ex = assertThrows<IllegalArgumentException> { service.createJournalEntry(request) }
        assertTrue(ex.message!!.contains("not balanced"))
        verify(journalEntryRepository, never()).save(any())
    }

    // ── 8. deactivateAccount — sets active=false ──────────────────────────────

    @Test
    fun `deactivateAccount sets active to false`() {
        val account = makeAccount().apply { active = true }
        whenever(accountRepository.findById(account.id)).thenReturn(Optional.of(account))
        whenever(accountRepository.save(any<Account>())).thenAnswer { it.arguments[0] as Account }

        val result = service.deactivateAccount(account.id)

        assertFalse(result.active)
        verify(accountRepository).save(argThat { !this.active })
    }

    // ── 9. getAccountBalance — returns current balance ────────────────────────

    @Test
    fun `getAccountBalance returns balance from account`() {
        val account = makeAccount(balance = BigDecimal("1234.56"))
        whenever(accountRepository.findById(account.id)).thenReturn(Optional.of(account))

        val balance = service.getAccountBalance(account.id)

        assertEquals(0, BigDecimal("1234.56").compareTo(balance))
    }

    // ── 10. balance update — debit increases ASSET balance ────────────────────

    @Test
    fun `createJournalEntry debit increases ASSET account balance`() {
        val cashAccount = makeAccount(type = AccountType.ASSET, code = "CASH", balance = BigDecimal("1000.00"))
        val expenseAccount = makeAccount(type = AccountType.EXPENSE, code = "OPEX", balance = BigDecimal("0.00"))

        val request = CreateJournalEntryRequest(
            reference = "EXP-001",
            description = "Expense payment",
            lines = listOf(
                JournalLineRequest(accountId = expenseAccount.id, type = "debit",  amount = BigDecimal("200.00")),
                JournalLineRequest(accountId = cashAccount.id,    type = "credit", amount = BigDecimal("200.00"))
            )
        )

        whenever(journalEntryRepository.existsByReference("EXP-001")).thenReturn(false)
        whenever(accountRepository.findAllById(any())).thenReturn(listOf(cashAccount, expenseAccount))

        val savedEntry = makeEntry().apply { reference = "EXP-001"; totalAmount = BigDecimal("200.00") }
        whenever(journalEntryRepository.save(any())).thenReturn(savedEntry)

        val savedAccounts = mutableListOf<Account>()
        whenever(accountRepository.save(any<Account>())).thenAnswer { invocation ->
            val a = invocation.arguments[0] as Account
            savedAccounts.add(a)
            a
        }

        service.createJournalEntry(request)

        // EXPENSE debited → balance increases by 200
        val savedExpense = savedAccounts.first { it.code == "OPEX" }
        assertEquals(0, BigDecimal("200.00").compareTo(savedExpense.balance))

        // ASSET credited → balance decreases by 200
        val savedCash = savedAccounts.first { it.code == "CASH" }
        assertEquals(0, BigDecimal("800.00").compareTo(savedCash.balance))
    }
}
