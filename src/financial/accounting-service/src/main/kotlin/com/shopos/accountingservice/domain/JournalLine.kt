package com.shopos.accountingservice.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.util.UUID

@Entity
@Table(name = "journal_lines")
class JournalLine {

    @Id
    @Column(name = "id", nullable = false, updatable = false)
    var id: UUID = UUID.randomUUID()

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "entry_id", nullable = false)
    var journalEntry: JournalEntry? = null

    @Column(name = "entry_id", insertable = false, updatable = false)
    var entryId: UUID = UUID.randomUUID()

    @Column(name = "account_id", nullable = false)
    var accountId: UUID = UUID.randomUUID()

    /**
     * "debit" or "credit"
     */
    @Column(name = "type", nullable = false, length = 10)
    var type: String = ""

    @Column(name = "amount", nullable = false, precision = 19, scale = 4)
    var amount: BigDecimal = BigDecimal.ZERO
}
