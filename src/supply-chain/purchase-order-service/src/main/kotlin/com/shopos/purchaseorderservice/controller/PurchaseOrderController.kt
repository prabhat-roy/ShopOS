package com.shopos.purchaseorderservice.controller

import com.shopos.purchaseorderservice.domain.POStatus
import com.shopos.purchaseorderservice.dto.CreatePORequest
import com.shopos.purchaseorderservice.dto.POResponse
import com.shopos.purchaseorderservice.dto.ReceiveItemsRequest
import com.shopos.purchaseorderservice.service.PurchaseOrderService
import jakarta.validation.Valid
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import java.util.UUID

@RestController
@RequestMapping("/purchase-orders")
class PurchaseOrderController(
    private val purchaseOrderService: PurchaseOrderService
) {

    @PostMapping
    fun createPO(@Valid @RequestBody request: CreatePORequest): ResponseEntity<POResponse> {
        val response = purchaseOrderService.createPO(request)
        return ResponseEntity.status(HttpStatus.CREATED).body(response)
    }

    @GetMapping("/{id}")
    fun getPO(@PathVariable id: UUID): ResponseEntity<POResponse> {
        val response = purchaseOrderService.getPO(id)
        return ResponseEntity.ok(response)
    }

    @GetMapping
    fun listPOs(
        @RequestParam(required = false) vendorId: UUID?,
        @RequestParam(required = false) status: POStatus?
    ): ResponseEntity<List<POResponse>> {
        val results = purchaseOrderService.listPOs(vendorId, status)
        return ResponseEntity.ok(results)
    }

    @PatchMapping("/{id}/submit")
    fun submitPO(@PathVariable id: UUID): ResponseEntity<POResponse> {
        val response = purchaseOrderService.submitPO(id)
        return ResponseEntity.ok(response)
    }

    @PatchMapping("/{id}/approve")
    fun approvePO(@PathVariable id: UUID): ResponseEntity<POResponse> {
        val response = purchaseOrderService.approvePO(id)
        return ResponseEntity.ok(response)
    }

    @PatchMapping("/{id}/reject")
    fun rejectPO(@PathVariable id: UUID): ResponseEntity<POResponse> {
        val response = purchaseOrderService.rejectPO(id)
        return ResponseEntity.ok(response)
    }

    @PostMapping("/{id}/receive")
    fun receiveItems(
        @PathVariable id: UUID,
        @Valid @RequestBody request: ReceiveItemsRequest
    ): ResponseEntity<POResponse> {
        val response = purchaseOrderService.receiveItems(id, request)
        return ResponseEntity.ok(response)
    }

    @DeleteMapping("/{id}")
    fun cancelPO(@PathVariable id: UUID): ResponseEntity<POResponse> {
        val response = purchaseOrderService.cancelPO(id)
        return ResponseEntity.ok(response)
    }
}
