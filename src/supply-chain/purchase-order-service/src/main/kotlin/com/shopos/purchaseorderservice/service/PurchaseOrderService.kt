package com.shopos.purchaseorderservice.service

import com.shopos.purchaseorderservice.domain.POStatus
import com.shopos.purchaseorderservice.domain.PurchaseOrder
import com.shopos.purchaseorderservice.domain.PurchaseOrderItem
import com.shopos.purchaseorderservice.dto.CreatePORequest
import com.shopos.purchaseorderservice.dto.POResponse
import com.shopos.purchaseorderservice.dto.ReceiveItemsRequest
import com.shopos.purchaseorderservice.exception.InvalidPOTransitionException
import com.shopos.purchaseorderservice.exception.PurchaseOrderNotFoundException
import com.shopos.purchaseorderservice.repository.PurchaseOrderRepository
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.util.UUID

@Service
class PurchaseOrderService(
    private val purchaseOrderRepository: PurchaseOrderRepository
) {

    @Transactional
    fun createPO(request: CreatePORequest): POResponse {
        val po = PurchaseOrder(
            vendorId = request.vendorId,
            status = POStatus.DRAFT,
            currency = request.currency,
            notes = request.notes,
            expectedDelivery = request.expectedDelivery
        )

        val items = request.items.map { itemReq ->
            val totalPrice = itemReq.unitPrice.multiply(BigDecimal(itemReq.quantity))
            PurchaseOrderItem(
                orderId = UUID.randomUUID(), // will be overwritten after save; managed via JoinColumn
                productId = itemReq.productId,
                sku = itemReq.sku,
                productName = itemReq.productName,
                quantity = itemReq.quantity,
                unitPrice = itemReq.unitPrice,
                totalPrice = totalPrice,
                receivedQty = 0
            )
        }

        val totalAmount = items.fold(BigDecimal.ZERO) { acc, item -> acc.add(item.totalPrice) }
        po.totalAmount = totalAmount
        po.items = items.toMutableList()

        val saved = purchaseOrderRepository.save(po)

        // Update orderId on each item to match the saved PO's id
        saved.items.forEach { it.orderId = saved.id!! }
        val finalSaved = purchaseOrderRepository.save(saved)

        return POResponse.from(finalSaved)
    }

    @Transactional(readOnly = true)
    fun getPO(id: UUID): POResponse {
        val po = purchaseOrderRepository.findById(id)
            .orElseThrow { PurchaseOrderNotFoundException(id) }
        return POResponse.from(po)
    }

    @Transactional(readOnly = true)
    fun listPOs(vendorId: UUID?, status: POStatus?): List<POResponse> {
        val results = when {
            vendorId != null && status != null ->
                purchaseOrderRepository.findByVendorIdAndStatus(vendorId, status)
            vendorId != null ->
                purchaseOrderRepository.findByVendorId(vendorId)
            status != null ->
                purchaseOrderRepository.findByStatus(status)
            else ->
                purchaseOrderRepository.findAll()
        }
        return results.map { POResponse.from(it) }
    }

    @Transactional
    fun submitPO(id: UUID): POResponse {
        val po = purchaseOrderRepository.findById(id)
            .orElseThrow { PurchaseOrderNotFoundException(id) }
        if (po.status != POStatus.DRAFT) {
            throw InvalidPOTransitionException(po.status, POStatus.SUBMITTED)
        }
        po.status = POStatus.SUBMITTED
        return POResponse.from(purchaseOrderRepository.save(po))
    }

    @Transactional
    fun approvePO(id: UUID): POResponse {
        val po = purchaseOrderRepository.findById(id)
            .orElseThrow { PurchaseOrderNotFoundException(id) }
        if (po.status != POStatus.SUBMITTED) {
            throw InvalidPOTransitionException(po.status, POStatus.APPROVED)
        }
        po.status = POStatus.APPROVED
        return POResponse.from(purchaseOrderRepository.save(po))
    }

    @Transactional
    fun rejectPO(id: UUID): POResponse {
        val po = purchaseOrderRepository.findById(id)
            .orElseThrow { PurchaseOrderNotFoundException(id) }
        if (po.status != POStatus.SUBMITTED) {
            throw InvalidPOTransitionException(po.status, POStatus.REJECTED)
        }
        po.status = POStatus.REJECTED
        return POResponse.from(purchaseOrderRepository.save(po))
    }

    @Transactional
    fun receiveItems(id: UUID, request: ReceiveItemsRequest): POResponse {
        val po = purchaseOrderRepository.findById(id)
            .orElseThrow { PurchaseOrderNotFoundException(id) }

        if (po.status != POStatus.APPROVED && po.status != POStatus.PARTIALLY_RECEIVED) {
            throw InvalidPOTransitionException(po.status, POStatus.PARTIALLY_RECEIVED)
        }

        val itemMap = po.items.associateBy { it.id }

        for (receipt in request.receipts) {
            val item = itemMap[receipt.itemId]
                ?: throw IllegalArgumentException("Item not found with id: ${receipt.itemId}")
            val newReceived = item.receivedQty + receipt.receivedQty
            if (newReceived > item.quantity) {
                throw IllegalArgumentException(
                    "Cannot receive ${receipt.receivedQty} units for item ${receipt.itemId}: " +
                    "would exceed ordered quantity of ${item.quantity} (already received: ${item.receivedQty})"
                )
            }
            item.receivedQty = newReceived
        }

        val allFullyReceived = po.items.all { it.receivedQty >= it.quantity }
        val anyReceived = po.items.any { it.receivedQty > 0 }

        po.status = when {
            allFullyReceived -> POStatus.FULLY_RECEIVED
            anyReceived -> POStatus.PARTIALLY_RECEIVED
            else -> po.status
        }

        return POResponse.from(purchaseOrderRepository.save(po))
    }

    @Transactional
    fun cancelPO(id: UUID): POResponse {
        val po = purchaseOrderRepository.findById(id)
            .orElseThrow { PurchaseOrderNotFoundException(id) }

        if (po.status == POStatus.FULLY_RECEIVED || po.status == POStatus.CANCELLED) {
            throw InvalidPOTransitionException(po.status, POStatus.CANCELLED)
        }
        po.status = POStatus.CANCELLED
        return POResponse.from(purchaseOrderRepository.save(po))
    }
}
