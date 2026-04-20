package com.shopos.reconciliationservice.dto

import com.shopos.reconciliationservice.domain.ReconciliationStatus
import java.math.BigDecimal

data class StatusCount(
    val status: ReconciliationStatus,
    val count: Long,
    val totalDiscrepancy: BigDecimal
)

data class ReconciliationSummary(
    val processor: String?,
    val totalRecords: Long,
    val totalDiscrepancy: BigDecimal,
    val byStatus: List<StatusCount>
)
