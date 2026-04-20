package com.shopos.orderservice.domain

import jakarta.persistence.*
import org.hibernate.annotations.CreationTimestamp
import org.hibernate.annotations.UpdateTimestamp
import java.math.BigDecimal
import java.time.OffsetDateTime
import java.util.UUID

@Entity
@Table(name = "orders")
data class Order(

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    val id: UUID = UUID.randomUUID(),

    @Column(name = "customer_id", nullable = false)
    val customerId: String = "",

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false)
    var status: OrderStatus = OrderStatus.PENDING,

    @OneToMany(cascade = [CascadeType.ALL], fetch = FetchType.EAGER, orphanRemoval = true)
    @JoinColumn(name = "order_id")
    val items: MutableList<OrderItem> = mutableListOf(),

    @Column(name = "subtotal", nullable = false, precision = 12, scale = 2)
    var subtotal: BigDecimal = BigDecimal.ZERO,

    @Column(name = "tax", nullable = false, precision = 12, scale = 2)
    var tax: BigDecimal = BigDecimal.ZERO,

    @Column(name = "shipping", nullable = false, precision = 12, scale = 2)
    var shipping: BigDecimal = BigDecimal.ZERO,

    @Column(name = "total", nullable = false, precision = 12, scale = 2)
    var total: BigDecimal = BigDecimal.ZERO,

    @Column(name = "currency", nullable = false, length = 3)
    val currency: String = "USD",

    // JSONB stored as text; serialized/deserialized at service layer
    @Column(name = "shipping_address", nullable = false, columnDefinition = "jsonb")
    var shippingAddress: String = "{}",

    @Column(name = "notes", nullable = false)
    var notes: String = "",

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    val createdAt: OffsetDateTime = OffsetDateTime.now(),

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    var updatedAt: OffsetDateTime = OffsetDateTime.now()
)
