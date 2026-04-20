package com.shopos.purchaseorderservice.dto

import com.shopos.purchaseorderservice.domain.POStatus
import com.shopos.purchaseorderservice.domain.PurchaseOrder
import com.shopos.purchaseorderservice.domain.PurchaseOrderItem
import java.math.BigDecimal
import java.time.Instant
import java.time.LocalDate
import java.util.UUID

data class POResponse(
    val id: UUID?,
    val vendorId: UUID,
    val status: POStatus,
    val totalAmount: BigDecimal,
    val currency: String,
    val notes: String?,
    val items: List<POItemResponse>,
    val expectedDelivery: LocalDate?,
    val createdAt: Instant?,
    val updatedAt: Instant?
) {
    companion object {
        fun from(po: PurchaseOrder): POResponse = POResponse(
            id = po.id,
            vendorId = po.vendorId,
            status = po.status,
            totalAmount = po.totalAmount,
            currency = po.currency,
            notes = po.notes,
            items = po.items.map { POItemResponse.from(it) },
            expectedDelivery = po.expectedDelivery,
            createdAt = po.createdAt,
            updatedAt = po.updatedAt
        )
    }
}

data class POItemResponse(
    val id: UUID?,
    val orderId: UUID,
    val productId: String,
    val sku: String,
    val productName: String,
    val quantity: Int,
    val unitPrice: BigDecimal,
    val totalPrice: BigDecimal,
    val receivedQty: Int
) {
    companion object {
        fun from(item: PurchaseOrderItem): POItemResponse = POItemResponse(
            id = item.id,
            orderId = item.orderId,
            productId = item.productId,
            sku = item.sku,
            productName = item.productName,
            quantity = item.quantity,
            unitPrice = item.unitPrice,
            totalPrice = item.totalPrice,
            receivedQty = item.receivedQty
        )
    }
}
