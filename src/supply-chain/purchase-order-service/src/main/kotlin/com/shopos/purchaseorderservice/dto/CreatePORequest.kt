package com.shopos.purchaseorderservice.dto

import jakarta.validation.Valid
import jakarta.validation.constraints.Future
import jakarta.validation.constraints.NotEmpty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Positive
import java.math.BigDecimal
import java.time.LocalDate
import java.util.UUID

data class CreatePORequest(

    @field:NotNull(message = "vendorId is required")
    val vendorId: UUID,

    @field:NotEmpty(message = "At least one item is required")
    @field:Valid
    val items: List<POItemRequest>,

    val notes: String? = null,

    @field:Future(message = "Expected delivery date must be in the future")
    val expectedDelivery: LocalDate? = null,

    val currency: String = "USD"
)

data class POItemRequest(

    @field:NotNull(message = "productId is required")
    val productId: String,

    @field:NotNull(message = "sku is required")
    val sku: String,

    @field:NotNull(message = "productName is required")
    val productName: String,

    @field:NotNull(message = "quantity is required")
    @field:Positive(message = "quantity must be positive")
    val quantity: Int,

    @field:NotNull(message = "unitPrice is required")
    @field:Positive(message = "unitPrice must be positive")
    val unitPrice: BigDecimal
)
