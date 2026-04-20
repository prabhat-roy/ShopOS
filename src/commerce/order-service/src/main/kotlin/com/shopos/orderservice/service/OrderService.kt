package com.shopos.orderservice.service

import com.shopos.orderservice.domain.Order
import com.shopos.orderservice.domain.OrderItem
import com.shopos.orderservice.domain.OrderStatus
import com.shopos.orderservice.dto.CreateOrderRequest
import com.shopos.orderservice.event.OrderEventPublisher
import com.shopos.orderservice.exception.NotFoundException
import com.shopos.orderservice.repository.OrderRepository
import org.springframework.data.domain.Page
import org.springframework.data.domain.Pageable
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.util.UUID

@Service
@Transactional
class OrderService(
    private val orderRepository: OrderRepository,
    private val eventPublisher: OrderEventPublisher
) {

    fun createOrder(req: CreateOrderRequest): Order {
        val items = req.items.map { itemReq ->
            OrderItem(
                productId = itemReq.productId,
                sku       = itemReq.sku,
                name      = itemReq.name,
                price     = itemReq.price,
                quantity  = itemReq.quantity,
                total     = itemReq.price.multiply(BigDecimal(itemReq.quantity))
            )
        }

        val subtotal = items.fold(BigDecimal.ZERO) { acc, item -> acc.add(item.total) }
        val total    = subtotal.add(req.tax).add(req.shipping)

        val order = Order(
            customerId      = req.customerId,
            status          = OrderStatus.PENDING,
            subtotal        = subtotal,
            tax             = req.tax,
            shipping        = req.shipping,
            total           = total,
            currency        = req.currency,
            shippingAddress = req.shippingAddress,
            notes           = req.notes
        )
        order.items.addAll(items)

        val saved = orderRepository.save(order)

        eventPublisher.publishOrderPlaced(
            saved.id,
            mapOf(
                "orderId"    to saved.id.toString(),
                "customerId" to saved.customerId,
                "total"      to saved.total,
                "currency"   to saved.currency,
                "status"     to saved.status.name
            )
        )

        return saved
    }

    @Transactional(readOnly = true)
    fun getOrder(id: UUID): Order =
        orderRepository.findById(id).orElseThrow {
            NotFoundException("Order not found: $id")
        }

    @Transactional(readOnly = true)
    fun listOrders(customerId: String?, status: OrderStatus?, pageable: Pageable): Page<Order> =
        when {
            customerId != null && status != null ->
                orderRepository.findByCustomerIdAndStatus(customerId, status, pageable)
            customerId != null ->
                orderRepository.findByCustomerId(customerId, pageable)
            status != null ->
                orderRepository.findByStatus(status, pageable)
            else ->
                orderRepository.findAll(pageable)
        }

    fun updateStatus(id: UUID, newStatus: OrderStatus): Order {
        val order = getOrder(id)
        order.status = newStatus
        val saved = orderRepository.save(order)

        if (newStatus == OrderStatus.CANCELLED) {
            eventPublisher.publishOrderCancelled(
                saved.id,
                mapOf(
                    "orderId"    to saved.id.toString(),
                    "customerId" to saved.customerId,
                    "status"     to saved.status.name
                )
            )
        }

        return saved
    }

    fun cancelOrder(id: UUID): Unit {
        updateStatus(id, OrderStatus.CANCELLED)
    }
}
