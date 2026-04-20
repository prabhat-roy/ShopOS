package com.shopos.reconciliationservice.dto

import jakarta.validation.constraints.DecimalMin
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Pattern
import java.math.BigDecimal
import java.util.UUID

data class ReconcileRequest(

    @field:NotNull(message = "internalPaymentId must not be null")
    val internalPaymentId: UUID,

    @field:NotBlank(message = "externalTransactionId must not be blank")
    val externalTransactionId: String,

    @field:NotNull(message = "internalAmount must not be null")
    @field:DecimalMin(value = "0.0001", message = "internalAmount must be greater than zero")
    val internalAmount: BigDecimal,

    @field:NotNull(message = "externalAmount must not be null")
    @field:DecimalMin(value = "0.0001", message = "externalAmount must be greater than zero")
    val externalAmount: BigDecimal,

    @field:Pattern(regexp = "^[A-Z]{3}$", message = "currency must be a 3-letter ISO code")
    val currency: String = "USD",

    @field:NotBlank(message = "processor must not be blank")
    val processor: String
)
