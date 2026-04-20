package com.shopos.revenuerecognition.dto

import com.shopos.revenuerecognition.domain.ContractType
import com.shopos.revenuerecognition.domain.RecognitionStatus
import java.math.BigDecimal
import java.time.Instant
import java.time.LocalDate
import java.util.UUID

data class ScheduleResponse(
    val id: UUID?,
    val orderId: String,
    val lineItemId: String,
    val contractType: ContractType,
    val totalAmount: BigDecimal,
    val recognizedAmount: BigDecimal,
    val deferredAmount: BigDecimal,
    val currency: String,
    val recognitionStartDate: LocalDate,
    val recognitionEndDate: LocalDate,
    val status: RecognitionStatus,
    val createdAt: Instant,
    val updatedAt: Instant
)
