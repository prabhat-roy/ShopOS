package com.shopos.revenuerecognition.dto

import com.shopos.revenuerecognition.domain.ContractType
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Positive
import java.math.BigDecimal
import java.time.LocalDate

data class CreateScheduleRequest(
    @field:NotBlank val orderId: String,
    @field:NotBlank val lineItemId: String,
    @field:NotNull val contractType: ContractType,
    @field:NotNull @field:Positive val totalAmount: BigDecimal,
    @field:NotBlank val currency: String,
    @field:NotNull val recognitionStartDate: LocalDate,
    @field:NotNull val recognitionEndDate: LocalDate
)
