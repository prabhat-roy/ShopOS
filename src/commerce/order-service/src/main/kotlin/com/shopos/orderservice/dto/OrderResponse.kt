package com.shopos.orderservice.dto

import com.shopos.orderservice.domain.Order
import com.shopos.orderservice.domain.OrderItem
import com.shopos.orderservice.domain.OrderStatus
import java.math.BigDecimal
import java.time.OffsetDateTime
import java.util.UUID

data class OrderResponse(
    val id: UUID,
    val customerId: String,
    val status: OrderStatus,
    val items: List<OrderItemResponse>,
    val subtotal: BigDecimal,
    val tax: BigDecimal,
    val shipping: BigDecimal,
    val total: BigDecimal,
    val currency: String,
    val shippingAddress: String,
    val notes: String,
    val createdAt: OffsetDateTime,
    val updatedAt: OffsetDateTime
) {
    companion object {
        fun from(order: Order) = OrderResponse(
            id              = order.id,
            customerId      = order.customerId,
            status          = order.status,
            items           = order.items.map { OrderItemResponse.from(it) },
            subtotal        = order.subtotal,
            tax             = order.tax,
            shipping        = order.shipping,
            total           = order.total,
            currency        = order.currency,
            shippingAddress = order.shippingAddress,
            notes           = order.notes,
            createdAt       = order.createdAt,
            updatedAt       = order.updatedAt
        )
    }
}

data class OrderItemResponse(
    val id: UUID,
    val productId: String,
    val sku: String,
    val name: String,
    val price: BigDecimal,
    val quantity: Int,
    val total: BigDecimal
) {
    companion object {
        fun from(item: OrderItem) = OrderItemResponse(
            id        = item.id,
            productId = item.productId,
            sku       = item.sku,
            name      = item.name,
            price     = item.price,
            quantity  = item.quantity,
            total     = item.total
        )
    }
}
