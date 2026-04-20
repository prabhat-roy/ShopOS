package com.shopos.purchaseorderservice.repository

import com.shopos.purchaseorderservice.domain.POStatus
import com.shopos.purchaseorderservice.domain.PurchaseOrder
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository
import java.util.UUID

@Repository
interface PurchaseOrderRepository : JpaRepository<PurchaseOrder, UUID> {

    fun findByVendorId(vendorId: UUID): List<PurchaseOrder>

    fun findByStatus(status: POStatus): List<PurchaseOrder>

    fun findByVendorIdAndStatus(vendorId: UUID, status: POStatus): List<PurchaseOrder>
}
