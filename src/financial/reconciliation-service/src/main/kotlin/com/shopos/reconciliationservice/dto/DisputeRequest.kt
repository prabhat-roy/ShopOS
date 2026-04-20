package com.shopos.reconciliationservice.dto

import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.Size

data class DisputeRequest(

    @field:NotBlank(message = "Dispute reason must not be blank")
    @field:Size(min = 10, max = 2000, message = "Dispute reason must be between 10 and 2000 characters")
    val reason: String
)
