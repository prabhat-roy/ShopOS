package com.shopos.contractservice.dto

import com.shopos.contractservice.domain.ContractType
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import java.math.BigDecimal
import java.time.LocalDate
import java.util.UUID

data class CreateContractRequest(

    @field:NotBlank(message = "title is required")
    val title: String,

    @field:NotNull(message = "orgId is required")
    val orgId: UUID,

    @field:NotNull(message = "type is required")
    val type: ContractType,

    @field:NotNull(message = "startDate is required")
    val startDate: LocalDate,

    @field:NotNull(message = "endDate is required")
    val endDate: LocalDate,

    @field:NotBlank(message = "createdBy is required")
    val createdBy: String,

    val vendorId: UUID? = null,
    val description: String? = null,
    val terms: String? = null,
    val value: BigDecimal? = null,
    val currency: String = "USD",
    val autoRenew: Boolean = false
)
