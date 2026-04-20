package com.shopos.subscriptionproduct.service

import com.shopos.subscriptionproduct.domain.SubscriptionPlan
import com.shopos.subscriptionproduct.dto.CreatePlanRequest
import com.shopos.subscriptionproduct.dto.UpdatePlanRequest
import com.shopos.subscriptionproduct.repository.SubscriptionPlanRepository
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.util.UUID

@Service
@Transactional
class SubscriptionPlanService(
    private val repository: SubscriptionPlanRepository
) {

    fun createPlan(request: CreatePlanRequest): SubscriptionPlan {
        val plan = SubscriptionPlan(
            productId    = request.productId,
            name         = request.name,
            description  = request.description,
            billingCycle = request.billingCycle,
            price        = request.price,
            currency     = request.currency,
            trialDays    = request.trialDays,
            active       = request.active,
            features     = request.features
        )
        return repository.save(plan)
    }

    @Transactional(readOnly = true)
    fun getAllPlans(): List<SubscriptionPlan> = repository.findAll()

    @Transactional(readOnly = true)
    fun getPlanById(id: UUID): SubscriptionPlan =
        repository.findById(id).orElseThrow {
            NoSuchElementException("Subscription plan not found: $id")
        }

    @Transactional(readOnly = true)
    fun getPlansByProductId(productId: String): List<SubscriptionPlan> =
        repository.findByProductId(productId)

    fun updatePlan(id: UUID, request: UpdatePlanRequest): SubscriptionPlan {
        val plan = getPlanById(id)
        request.name?.let         { plan.name         = it }
        request.description?.let  { plan.description  = it }
        request.billingCycle?.let { plan.billingCycle = it }
        request.price?.let        { plan.price        = it }
        request.currency?.let     { plan.currency     = it }
        request.trialDays?.let    { plan.trialDays    = it }
        request.active?.let       { plan.active       = it }
        request.features?.let     { plan.features     = it }
        return repository.save(plan)
    }

    fun deletePlan(id: UUID) {
        if (!repository.existsById(id)) {
            throw NoSuchElementException("Subscription plan not found: $id")
        }
        repository.deleteById(id)
    }
}
