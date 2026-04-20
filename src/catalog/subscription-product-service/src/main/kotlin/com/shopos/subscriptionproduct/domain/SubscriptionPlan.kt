package com.shopos.subscriptionproduct.domain

import jakarta.persistence.*
import org.hibernate.annotations.CreationTimestamp
import org.hibernate.annotations.UpdateTimestamp
import java.math.BigDecimal
import java.time.OffsetDateTime
import java.util.UUID

enum class BillingCycle {
    MONTHLY, QUARTERLY, ANNUAL
}

@Entity
@Table(
    name = "subscription_plans",
    indexes = [Index(name = "idx_sub_plans_product_id", columnList = "product_id")]
)
data class SubscriptionPlan(

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    val id: UUID = UUID.randomUUID(),

    @Column(name = "product_id", nullable = false)
    var productId: String,

    @Column(name = "name", nullable = false)
    var name: String,

    @Column(name = "description", nullable = false)
    var description: String = "",

    @Enumerated(EnumType.STRING)
    @Column(name = "billing_cycle", nullable = false)
    var billingCycle: BillingCycle = BillingCycle.MONTHLY,

    @Column(name = "price", nullable = false, precision = 12, scale = 2)
    var price: BigDecimal,

    @Column(name = "currency", nullable = false, length = 3)
    var currency: String = "USD",

    @Column(name = "trial_days", nullable = false)
    var trialDays: Int = 0,

    @Column(name = "active", nullable = false)
    var active: Boolean = true,

    @Column(name = "features", columnDefinition = "TEXT[]")
    @Convert(converter = StringArrayConverter::class)
    var features: List<String> = emptyList(),

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    val createdAt: OffsetDateTime = OffsetDateTime.now(),

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    var updatedAt: OffsetDateTime = OffsetDateTime.now()
)

@Converter(autoApply = false)
class StringArrayConverter : AttributeConverter<List<String>, Array<String>> {
    override fun convertToDatabaseColumn(attribute: List<String>?): Array<String> =
        attribute?.toTypedArray() ?: emptyArray()

    override fun convertToEntityAttribute(dbData: Array<String>?): List<String> =
        dbData?.toList() ?: emptyList()
}
