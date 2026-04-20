package com.shopos.accountingservice.dto

import jakarta.validation.Valid
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotEmpty
import jakarta.validation.constraints.Pattern
import jakarta.validation.constraints.Size

data class CreateJournalEntryRequest(

    @field:NotBlank(message = "Reference must not be blank")
    @field:Size(min = 2, max = 100, message = "Reference must be between 2 and 100 characters")
    val reference: String,

    @field:NotBlank(message = "Description must not be blank")
    @field:Size(max = 500, message = "Description must not exceed 500 characters")
    val description: String,

    @field:Pattern(regexp = "^[A-Z]{3}$", message = "Currency must be a 3-letter ISO code")
    val currency: String = "USD",

    @field:NotEmpty(message = "Journal entry must have at least one line")
    @field:Valid
    val lines: List<JournalLineRequest>
)
