package com.shopos.accountingservice.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

@Entity
@Table(name = "journal_entries")
class JournalEntry {

    @Id
    @Column(name = "id", nullable = false, updatable = false)
    var id: UUID = UUID.randomUUID()

    @Column(name = "reference", nullable = false, unique = true, length = 100)
    var reference: String = ""

    @Column(name = "description", nullable = false, length = 500)
    var description: String = ""

    @Column(name = "total_amount", nullable = false, precision = 19, scale = 4)
    var totalAmount: BigDecimal = BigDecimal.ZERO

    @Column(name = "currency", nullable = false, length = 3)
    var currency: String = "USD"

    @OneToMany(
        mappedBy = "journalEntry",
        cascade = [CascadeType.ALL],
        orphanRemoval = true,
        fetch = FetchType.EAGER
    )
    var lines: MutableList<JournalLine> = mutableListOf()

    @Column(name = "created_at", nullable = false, updatable = false)
    var createdAt: LocalDateTime = LocalDateTime.now()
}
