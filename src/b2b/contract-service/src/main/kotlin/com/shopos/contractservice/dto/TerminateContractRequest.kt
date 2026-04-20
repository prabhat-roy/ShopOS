package com.shopos.contractservice.dto

import jakarta.validation.constraints.NotBlank

data class TerminateContractRequest(
    @field:NotBlank(message = "reason is required")
    val reason: String
)
