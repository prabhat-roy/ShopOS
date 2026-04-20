package com.shopos.accountingservice.repository

import com.shopos.accountingservice.domain.JournalEntry
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Query
import org.springframework.data.repository.query.Param
import org.springframework.stereotype.Repository
import java.time.LocalDateTime
import java.util.Optional
import java.util.UUID

@Repository
interface JournalEntryRepository : JpaRepository<JournalEntry, UUID> {

    fun findByReference(reference: String): Optional<JournalEntry>

    fun findByCreatedAtBetween(start: LocalDateTime, end: LocalDateTime): List<JournalEntry>

    fun existsByReference(reference: String): Boolean

    @Query("""
        SELECT DISTINCT je FROM JournalEntry je
        JOIN je.lines jl
        WHERE jl.accountId = :accountId
        AND je.createdAt BETWEEN :start AND :end
        ORDER BY je.createdAt ASC
    """)
    fun findByAccountIdAndDateRange(
        @Param("accountId") accountId: UUID,
        @Param("start") start: LocalDateTime,
        @Param("end") end: LocalDateTime
    ): List<JournalEntry>
}
