package com.shopos.accountingservice.dto

import com.shopos.accountingservice.domain.JournalEntry
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

data class JournalLineResponse(
    val id: UUID,
    val entryId: UUID,
    val accountId: UUID,
    val type: String,
    val amount: BigDecimal
)

data class JournalEntryResponse(
    val id: UUID,
    val reference: String,
    val description: String,
    val totalAmount: BigDecimal,
    val currency: String,
    val lines: List<JournalLineResponse>,
    val createdAt: LocalDateTime
) {
    companion object {
        fun from(entry: JournalEntry): JournalEntryResponse = JournalEntryResponse(
            id = entry.id,
            reference = entry.reference,
            description = entry.description,
            totalAmount = entry.totalAmount,
            currency = entry.currency,
            lines = entry.lines.map { line ->
                JournalLineResponse(
                    id = line.id,
                    entryId = entry.id,
                    accountId = line.accountId,
                    type = line.type,
                    amount = line.amount
                )
            },
            createdAt = entry.createdAt
        )
    }
}
