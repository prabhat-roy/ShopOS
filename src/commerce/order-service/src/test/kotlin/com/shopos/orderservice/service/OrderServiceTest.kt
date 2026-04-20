package com.shopos.orderservice.service

import com.shopos.orderservice.domain.Order
import com.shopos.orderservice.domain.OrderItem
import com.shopos.orderservice.domain.OrderStatus
import com.shopos.orderservice.dto.CreateOrderRequest
import com.shopos.orderservice.dto.OrderItemRequest
import com.shopos.orderservice.event.OrderEventPublisher
import com.shopos.orderservice.exception.NotFoundException
import com.shopos.orderservice.repository.OrderRepository
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.*
import java.math.BigDecimal
import java.util.Optional
import java.util.UUID

class OrderServiceTest {

    private val orderRepository: OrderRepository   = mock()
    private val eventPublisher: OrderEventPublisher = mock()
    private val service = OrderService(orderRepository, eventPublisher)

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------

    private fun buildOrder(
        id: UUID         = UUID.randomUUID(),
        customerId: String = "cust-001",
        status: OrderStatus = OrderStatus.PENDING
    ): Order {
        val item = OrderItem(
            orderId   = id,
            productId = "prod-001",
            sku       = "SKU-001",
            name      = "Widget",
            price     = BigDecimal("10.00"),
            quantity  = 2,
            total     = BigDecimal("20.00")
        )
        return Order(
            id         = id,
            customerId = customerId,
            status     = status,
            subtotal   = BigDecimal("20.00"),
            tax        = BigDecimal("2.00"),
            shipping   = BigDecimal("5.00"),
            total      = BigDecimal("27.00"),
            currency   = "USD"
        ).also { it.items.add(item) }
    }

    private fun buildCreateRequest(customerId: String = "cust-001") = CreateOrderRequest(
        customerId = customerId,
        items = listOf(
            OrderItemRequest(
                productId = "prod-001",
                sku       = "SKU-001",
                name      = "Widget",
                price     = BigDecimal("10.00"),
                quantity  = 2
            )
        ),
        tax      = BigDecimal("2.00"),
        shipping = BigDecimal("5.00")
    )

    // -------------------------------------------------------------------------
    // createOrder
    // -------------------------------------------------------------------------

    @Test
    fun `createOrder saves order and publishes commerce_order_placed event`() {
        val req   = buildCreateRequest()
        val saved = buildOrder(customerId = req.customerId)

        whenever(orderRepository.save(any<Order>())).thenReturn(saved)

        val result = service.createOrder(req)

        verify(orderRepository).save(any<Order>())
        verify(eventPublisher).publishOrderPlaced(eq(saved.id), any())
        assertEquals(saved.id, result.id)
        assertEquals("cust-001", result.customerId)
        assertEquals(OrderStatus.PENDING, result.status)
    }

    @Test
    fun `createOrder computes subtotal and total from items`() {
        val req = buildCreateRequest()
        val capturedOrder = argumentCaptor<Order>()

        whenever(orderRepository.save(capturedOrder.capture())).thenAnswer {
            capturedOrder.firstValue
        }

        service.createOrder(req)

        val saved = capturedOrder.firstValue
        // subtotal = 10.00 * 2 = 20.00
        assertEquals(BigDecimal("20.00"), saved.subtotal)
        // total = 20.00 + 2.00 (tax) + 5.00 (shipping)
        assertEquals(BigDecimal("27.00"), saved.total)
    }

    // -------------------------------------------------------------------------
    // getOrder
    // -------------------------------------------------------------------------

    @Test
    fun `getOrder returns order when found`() {
        val id    = UUID.randomUUID()
        val order = buildOrder(id = id)
        whenever(orderRepository.findById(id)).thenReturn(Optional.of(order))

        val result = service.getOrder(id)

        assertEquals(id, result.id)
    }

    @Test
    fun `getOrder throws NotFoundException when order does not exist`() {
        val id = UUID.randomUUID()
        whenever(orderRepository.findById(id)).thenReturn(Optional.empty())

        assertThrows<NotFoundException> {
            service.getOrder(id)
        }
    }

    // -------------------------------------------------------------------------
    // updateStatus
    // -------------------------------------------------------------------------

    @Test
    fun `updateStatus to CANCELLED publishes commerce_order_cancelled event`() {
        val id    = UUID.randomUUID()
        val order = buildOrder(id = id, status = OrderStatus.CONFIRMED)
        whenever(orderRepository.findById(id)).thenReturn(Optional.of(order))
        whenever(orderRepository.save(any<Order>())).thenAnswer { it.arguments[0] }

        service.updateStatus(id, OrderStatus.CANCELLED)

        verify(eventPublisher).publishOrderCancelled(eq(id), any())
        assertEquals(OrderStatus.CANCELLED, order.status)
    }

    @Test
    fun `updateStatus to non-CANCELLED does not publish cancel event`() {
        val id    = UUID.randomUUID()
        val order = buildOrder(id = id, status = OrderStatus.PENDING)
        whenever(orderRepository.findById(id)).thenReturn(Optional.of(order))
        whenever(orderRepository.save(any<Order>())).thenAnswer { it.arguments[0] }

        service.updateStatus(id, OrderStatus.CONFIRMED)

        verify(eventPublisher, never()).publishOrderCancelled(any(), any())
        assertEquals(OrderStatus.CONFIRMED, order.status)
    }

    @Test
    fun `updateStatus throws NotFoundException when order not found`() {
        val id = UUID.randomUUID()
        whenever(orderRepository.findById(id)).thenReturn(Optional.empty())

        assertThrows<NotFoundException> {
            service.updateStatus(id, OrderStatus.CANCELLED)
        }
    }
}
