package com.shopos.subscriptionproduct.controller

import com.shopos.subscriptionproduct.dto.CreatePlanRequest
import com.shopos.subscriptionproduct.dto.PlanResponse
import com.shopos.subscriptionproduct.dto.UpdatePlanRequest
import com.shopos.subscriptionproduct.service.SubscriptionPlanService
import jakarta.validation.Valid
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import java.util.UUID

@RestController
@RequestMapping("/plans")
class SubscriptionPlanController(
    private val service: SubscriptionPlanService
) {

    @PostMapping
    fun createPlan(@Valid @RequestBody request: CreatePlanRequest): ResponseEntity<PlanResponse> {
        val plan = service.createPlan(request)
        return ResponseEntity.status(HttpStatus.CREATED).body(PlanResponse.from(plan))
    }

    @GetMapping
    fun getAllPlans(): ResponseEntity<List<PlanResponse>> {
        val plans = service.getAllPlans().map { PlanResponse.from(it) }
        return ResponseEntity.ok(plans)
    }

    @GetMapping("/{id}")
    fun getPlanById(@PathVariable id: UUID): ResponseEntity<PlanResponse> {
        val plan = service.getPlanById(id)
        return ResponseEntity.ok(PlanResponse.from(plan))
    }

    @GetMapping("/product/{productId}")
    fun getPlansByProductId(@PathVariable productId: String): ResponseEntity<List<PlanResponse>> {
        val plans = service.getPlansByProductId(productId).map { PlanResponse.from(it) }
        return ResponseEntity.ok(plans)
    }

    @PatchMapping("/{id}")
    fun updatePlan(
        @PathVariable id: UUID,
        @Valid @RequestBody request: UpdatePlanRequest
    ): ResponseEntity<PlanResponse> {
        val plan = service.updatePlan(id, request)
        return ResponseEntity.ok(PlanResponse.from(plan))
    }

    @DeleteMapping("/{id}")
    fun deletePlan(@PathVariable id: UUID): ResponseEntity<Void> {
        service.deletePlan(id)
        return ResponseEntity.noContent().build()
    }
}
