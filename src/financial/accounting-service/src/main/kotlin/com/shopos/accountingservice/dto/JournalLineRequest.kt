package com.shopos.accountingservice.dto

import jakarta.validation.constraints.DecimalMin
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Pattern
import java.math.BigDecimal
import java.util.UUID

data class JournalLineRequest(

    @field:NotNull(message = "Account ID must not be null")
    val accountId: UUID,

    @field:NotNull(message = "Line type must not be null")
    @field:Pattern(regexp = "^(debit|credit)$", message = "Line type must be 'debit' or 'credit'")
    val type: String,

    @field:NotNull(message = "Amount must not be null")
    @field:DecimalMin(value = "0.0001", message = "Amount must be greater than zero")
    val amount: BigDecimal
)
