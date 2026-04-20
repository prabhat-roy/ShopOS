package com.shopos.orderservice.dto

import jakarta.validation.Valid
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotEmpty
import jakarta.validation.constraints.Positive
import java.math.BigDecimal

data class CreateOrderRequest(

    @field:NotBlank(message = "customerId is required")
    val customerId: String,

    @field:NotEmpty(message = "items must not be empty")
    @field:Valid
    val items: List<OrderItemRequest>,

    val tax: BigDecimal = BigDecimal.ZERO,
    val shipping: BigDecimal = BigDecimal.ZERO,
    val currency: String = "USD",
    val shippingAddress: String = "{}",
    val notes: String = ""
)

data class OrderItemRequest(

    @field:NotBlank(message = "productId is required")
    val productId: String,

    val sku: String = "",

    @field:NotBlank(message = "name is required")
    val name: String,

    @field:Positive(message = "price must be positive")
    val price: BigDecimal,

    @field:Positive(message = "quantity must be positive")
    val quantity: Int
)
