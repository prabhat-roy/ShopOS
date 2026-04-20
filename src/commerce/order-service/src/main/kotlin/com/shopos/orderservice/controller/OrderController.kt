package com.shopos.orderservice.controller

import com.shopos.orderservice.domain.OrderStatus
import com.shopos.orderservice.dto.CreateOrderRequest
import com.shopos.orderservice.dto.OrderResponse
import com.shopos.orderservice.service.OrderService
import jakarta.validation.Valid
import org.springframework.data.domain.Page
import org.springframework.data.domain.PageRequest
import org.springframework.data.domain.Sort
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import java.util.UUID

@RestController
@RequestMapping("/orders")
class OrderController(
    private val orderService: OrderService
) {

    @PostMapping
    fun createOrder(
        @Valid @RequestBody req: CreateOrderRequest
    ): ResponseEntity<OrderResponse> {
        val order = orderService.createOrder(req)
        return ResponseEntity.status(HttpStatus.CREATED).body(OrderResponse.from(order))
    }

    @GetMapping("/{id}")
    fun getOrder(@PathVariable id: UUID): ResponseEntity<OrderResponse> {
        val order = orderService.getOrder(id)
        return ResponseEntity.ok(OrderResponse.from(order))
    }

    @GetMapping
    fun listOrders(
        @RequestParam(required = false) customerId: String?,
        @RequestParam(required = false) status: OrderStatus?,
        @RequestParam(defaultValue = "0") page: Int,
        @RequestParam(defaultValue = "20") size: Int
    ): ResponseEntity<Page<OrderResponse>> {
        val pageable = PageRequest.of(page, size, Sort.by("createdAt").descending())
        val orders   = orderService.listOrders(customerId, status, pageable)
        return ResponseEntity.ok(orders.map { OrderResponse.from(it) })
    }

    @PatchMapping("/{id}/status")
    fun updateStatus(
        @PathVariable id: UUID,
        @RequestBody body: Map<String, String>
    ): ResponseEntity<OrderResponse> {
        val newStatus = OrderStatus.valueOf(
            body["status"] ?: throw IllegalArgumentException("status field is required")
        )
        val order = orderService.updateStatus(id, newStatus)
        return ResponseEntity.ok(OrderResponse.from(order))
    }

    @DeleteMapping("/{id}")
    fun cancelOrder(@PathVariable id: UUID): ResponseEntity<Void> {
        orderService.cancelOrder(id)
        return ResponseEntity.noContent().build()
    }
}
