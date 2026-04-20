package com.shopos.orderservice.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.util.UUID

@Entity
@Table(name = "order_items")
data class OrderItem(

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    val id: UUID = UUID.randomUUID(),

    @Column(name = "order_id", nullable = false)
    val orderId: UUID = UUID.randomUUID(),

    @Column(name = "product_id", nullable = false)
    val productId: String = "",

    @Column(name = "sku", nullable = false)
    val sku: String = "",

    @Column(name = "name", nullable = false)
    val name: String = "",

    @Column(name = "price", nullable = false, precision = 12, scale = 2)
    val price: BigDecimal = BigDecimal.ZERO,

    @Column(name = "quantity", nullable = false)
    val quantity: Int = 1,

    @Column(name = "total", nullable = false, precision = 12, scale = 2)
    val total: BigDecimal = BigDecimal.ZERO
)
