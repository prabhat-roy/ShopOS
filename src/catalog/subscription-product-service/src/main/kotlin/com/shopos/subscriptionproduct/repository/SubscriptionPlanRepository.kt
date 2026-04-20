package com.shopos.subscriptionproduct.repository

import com.shopos.subscriptionproduct.domain.SubscriptionPlan
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository
import java.util.UUID

@Repository
interface SubscriptionPlanRepository : JpaRepository<SubscriptionPlan, UUID> {
    fun findByProductId(productId: String): List<SubscriptionPlan>
    fun findByProductIdAndActiveTrue(productId: String): List<SubscriptionPlan>
}
