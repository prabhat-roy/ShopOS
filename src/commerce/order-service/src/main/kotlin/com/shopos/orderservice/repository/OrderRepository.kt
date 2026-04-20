package com.shopos.orderservice.repository

import com.shopos.orderservice.domain.Order
import com.shopos.orderservice.domain.OrderStatus
import org.springframework.data.domain.Page
import org.springframework.data.domain.Pageable
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository
import java.util.UUID

@Repository
interface OrderRepository : JpaRepository<Order, UUID> {

    fun findByCustomerId(customerId: String, pageable: Pageable): Page<Order>

    fun findByStatus(status: OrderStatus, pageable: Pageable): Page<Order>

    fun findByCustomerIdAndStatus(customerId: String, status: OrderStatus, pageable: Pageable): Page<Order>
}
