package com.shopos.subscriptionproduct.service

import com.shopos.subscriptionproduct.domain.BillingCycle
import com.shopos.subscriptionproduct.domain.SubscriptionPlan
import com.shopos.subscriptionproduct.dto.CreatePlanRequest
import com.shopos.subscriptionproduct.dto.UpdatePlanRequest
import com.shopos.subscriptionproduct.repository.SubscriptionPlanRepository
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.*
import java.math.BigDecimal
import java.util.*

class SubscriptionPlanServiceTest {

    private val repository: SubscriptionPlanRepository = mock()
    private val service = SubscriptionPlanService(repository)

    private fun buildPlan(
        id: UUID = UUID.randomUUID(),
        productId: String = "prod-001",
        name: String = "Monthly Plan",
        price: BigDecimal = BigDecimal("9.99"),
        billingCycle: BillingCycle = BillingCycle.MONTHLY
    ) = SubscriptionPlan(
        id           = id,
        productId    = productId,
        name         = name,
        description  = "Test plan",
        billingCycle = billingCycle,
        price        = price,
        currency     = "USD",
        trialDays    = 7,
        active       = true,
        features     = listOf("feature-a", "feature-b")
    )

    // -------------------------------------------------------------------------
    // createPlan
    // -------------------------------------------------------------------------

    @Test
    fun `createPlan saves and returns plan`() {
        val request = CreatePlanRequest(
            productId    = "prod-001",
            name         = "Monthly Plan",
            description  = "A monthly subscription",
            billingCycle = BillingCycle.MONTHLY,
            price        = BigDecimal("9.99"),
            currency     = "USD",
            trialDays    = 7,
            active       = true,
            features     = listOf("feature-a")
        )
        val savedPlan = buildPlan(productId = request.productId, name = request.name)
        whenever(repository.save(any<SubscriptionPlan>())).thenReturn(savedPlan)

        val result = service.createPlan(request)

        verify(repository).save(any<SubscriptionPlan>())
        assertEquals(savedPlan.id, result.id)
        assertEquals("prod-001", result.productId)
        assertEquals("Monthly Plan", result.name)
        assertEquals(BillingCycle.MONTHLY, result.billingCycle)
    }

    @Test
    fun `createPlan with annual billing cycle is persisted correctly`() {
        val request = CreatePlanRequest(
            productId    = "prod-002",
            name         = "Annual Plan",
            billingCycle = BillingCycle.ANNUAL,
            price        = BigDecimal("99.99")
        )
        val savedPlan = buildPlan(
            productId    = request.productId,
            name         = request.name,
            billingCycle = BillingCycle.ANNUAL,
            price        = request.price
        )
        whenever(repository.save(any<SubscriptionPlan>())).thenReturn(savedPlan)

        val result = service.createPlan(request)

        assertEquals(BillingCycle.ANNUAL, result.billingCycle)
        assertEquals(BigDecimal("99.99"), result.price)
    }

    // -------------------------------------------------------------------------
    // getPlansByProductId
    // -------------------------------------------------------------------------

    @Test
    fun `getByProductId returns all plans for a product`() {
        val productId = "prod-001"
        val plans = listOf(
            buildPlan(productId = productId, name = "Monthly Plan", billingCycle = BillingCycle.MONTHLY),
            buildPlan(productId = productId, name = "Annual Plan",  billingCycle = BillingCycle.ANNUAL)
        )
        whenever(repository.findByProductId(productId)).thenReturn(plans)

        val result = service.getPlansByProductId(productId)

        verify(repository).findByProductId(productId)
        assertEquals(2, result.size)
        assertTrue(result.all { it.productId == productId })
    }

    @Test
    fun `getByProductId returns empty list when no plans exist`() {
        whenever(repository.findByProductId("unknown-prod")).thenReturn(emptyList())

        val result = service.getPlansByProductId("unknown-prod")

        assertTrue(result.isEmpty())
    }

    // -------------------------------------------------------------------------
    // deletePlan
    // -------------------------------------------------------------------------

    @Test
    fun `deletePlan removes existing plan`() {
        val id = UUID.randomUUID()
        whenever(repository.existsById(id)).thenReturn(true)

        service.deletePlan(id)

        verify(repository).deleteById(id)
    }

    @Test
    fun `deletePlan throws NoSuchElementException when plan not found`() {
        val id = UUID.randomUUID()
        whenever(repository.existsById(id)).thenReturn(false)

        assertThrows<NoSuchElementException> {
            service.deletePlan(id)
        }
        verify(repository, never()).deleteById(any())
    }

    // -------------------------------------------------------------------------
    // updatePlan
    // -------------------------------------------------------------------------

    @Test
    fun `updatePlan patches only provided fields`() {
        val id = UUID.randomUUID()
        val existing = buildPlan(id = id, name = "Old Name", price = BigDecimal("9.99"))
        val updateRequest = UpdatePlanRequest(name = "New Name", price = BigDecimal("14.99"))

        whenever(repository.findById(id)).thenReturn(Optional.of(existing))
        whenever(repository.save(any<SubscriptionPlan>())).thenAnswer { it.arguments[0] }

        val result = service.updatePlan(id, updateRequest)

        assertEquals("New Name", result.name)
        assertEquals(BigDecimal("14.99"), result.price)
        // Fields not in the patch request stay unchanged
        assertEquals("prod-001", result.productId)
        assertEquals(BillingCycle.MONTHLY, result.billingCycle)
    }

    @Test
    fun `updatePlan throws NoSuchElementException when plan not found`() {
        val id = UUID.randomUUID()
        whenever(repository.findById(id)).thenReturn(Optional.empty())

        assertThrows<NoSuchElementException> {
            service.updatePlan(id, UpdatePlanRequest(name = "X"))
        }
    }

    // -------------------------------------------------------------------------
    // getPlanById
    // -------------------------------------------------------------------------

    @Test
    fun `getPlanById returns plan when found`() {
        val id = UUID.randomUUID()
        val plan = buildPlan(id = id)
        whenever(repository.findById(id)).thenReturn(Optional.of(plan))

        val result = service.getPlanById(id)

        assertEquals(id, result.id)
    }

    @Test
    fun `getPlanById throws NoSuchElementException when not found`() {
        val id = UUID.randomUUID()
        whenever(repository.findById(id)).thenReturn(Optional.empty())

        assertThrows<NoSuchElementException> {
            service.getPlanById(id)
        }
    }
}
