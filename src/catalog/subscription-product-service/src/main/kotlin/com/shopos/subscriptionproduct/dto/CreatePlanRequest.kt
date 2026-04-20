package com.shopos.subscriptionproduct.dto

import com.shopos.subscriptionproduct.domain.BillingCycle
import jakarta.validation.constraints.*
import java.math.BigDecimal

data class CreatePlanRequest(

    @field:NotBlank(message = "productId must not be blank")
    val productId: String,

    @field:NotBlank(message = "name must not be blank")
    @field:Size(max = 255, message = "name must be 255 characters or fewer")
    val name: String,

    val description: String = "",

    val billingCycle: BillingCycle = BillingCycle.MONTHLY,

    @field:NotNull(message = "price is required")
    @field:DecimalMin(value = "0.00", inclusive = false, message = "price must be greater than 0")
    @field:Digits(integer = 10, fraction = 2, message = "price must have at most 10 integer digits and 2 decimal places")
    val price: BigDecimal,

    @field:Size(min = 3, max = 3, message = "currency must be a 3-letter ISO code")
    val currency: String = "USD",

    @field:Min(value = 0, message = "trialDays must be 0 or greater")
    val trialDays: Int = 0,

    val active: Boolean = true,

    val features: List<String> = emptyList()
)

data class UpdatePlanRequest(
    val name: String? = null,
    val description: String? = null,
    val billingCycle: BillingCycle? = null,
    @field:DecimalMin(value = "0.00", inclusive = false, message = "price must be greater than 0")
    @field:Digits(integer = 10, fraction = 2)
    val price: BigDecimal? = null,
    @field:Size(min = 3, max = 3)
    val currency: String? = null,
    @field:Min(0)
    val trialDays: Int? = null,
    val active: Boolean? = null,
    val features: List<String>? = null
)
