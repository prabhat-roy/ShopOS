package com.shopos.subscriptionproduct.dto

import com.shopos.subscriptionproduct.domain.BillingCycle
import com.shopos.subscriptionproduct.domain.SubscriptionPlan
import java.math.BigDecimal
import java.time.OffsetDateTime
import java.util.UUID

data class PlanResponse(
    val id: UUID,
    val productId: String,
    val name: String,
    val description: String,
    val billingCycle: BillingCycle,
    val price: BigDecimal,
    val currency: String,
    val trialDays: Int,
    val active: Boolean,
    val features: List<String>,
    val createdAt: OffsetDateTime,
    val updatedAt: OffsetDateTime
) {
    companion object {
        fun from(plan: SubscriptionPlan) = PlanResponse(
            id           = plan.id,
            productId    = plan.productId,
            name         = plan.name,
            description  = plan.description,
            billingCycle = plan.billingCycle,
            price        = plan.price,
            currency     = plan.currency,
            trialDays    = plan.trialDays,
            active       = plan.active,
            features     = plan.features,
            createdAt    = plan.createdAt,
            updatedAt    = plan.updatedAt
        )
    }
}
