package com.shopos.accountingservice.dto

import com.shopos.accountingservice.domain.AccountType
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Pattern
import jakarta.validation.constraints.Size

data class CreateAccountRequest(

    @field:NotBlank(message = "Account code must not be blank")
    @field:Size(min = 2, max = 50, message = "Account code must be between 2 and 50 characters")
    @field:Pattern(regexp = "^[A-Z0-9_-]+$", message = "Account code must be uppercase alphanumeric with hyphens/underscores only")
    val code: String,

    @field:NotBlank(message = "Account name must not be blank")
    @field:Size(min = 2, max = 255, message = "Account name must be between 2 and 255 characters")
    val name: String,

    @field:NotNull(message = "Account type must not be null")
    val type: AccountType,

    @field:Pattern(regexp = "^[A-Z]{3}$", message = "Currency must be a 3-letter ISO code")
    val currency: String = "USD"
)
