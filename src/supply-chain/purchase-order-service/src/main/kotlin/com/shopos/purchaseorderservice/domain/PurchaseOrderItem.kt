package com.shopos.purchaseorderservice.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.util.UUID

@Entity
@Table(name = "purchase_order_items")
class PurchaseOrderItem(

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    var id: UUID? = null,

    @Column(name = "order_id", nullable = false)
    var orderId: UUID,

    @Column(name = "product_id", nullable = false, length = 255)
    var productId: String,

    @Column(name = "sku", nullable = false, length = 100)
    var sku: String,

    @Column(name = "product_name", nullable = false, length = 255)
    var productName: String,

    @Column(name = "quantity", nullable = false)
    var quantity: Int,

    @Column(name = "unit_price", nullable = false, precision = 19, scale = 4)
    var unitPrice: BigDecimal,

    @Column(name = "total_price", nullable = false, precision = 19, scale = 4)
    var totalPrice: BigDecimal,

    @Column(name = "received_qty", nullable = false)
    var receivedQty: Int = 0
)
