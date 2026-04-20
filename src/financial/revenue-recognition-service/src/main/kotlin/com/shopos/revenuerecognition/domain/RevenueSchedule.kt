package com.shopos.revenuerecognition.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.time.Instant
import java.time.LocalDate
import java.util.UUID

@Entity
@Table(name = "revenue_schedules")
data class RevenueSchedule(

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    val id: UUID? = null,

    @Column(name = "order_id", nullable = false)
    val orderId: String,

    @Column(name = "line_item_id", nullable = false)
    val lineItemId: String,

    @Column(name = "contract_type", nullable = false)
    @Enumerated(EnumType.STRING)
    val contractType: ContractType,

    @Column(name = "total_amount", nullable = false, precision = 19, scale = 4)
    val totalAmount: BigDecimal,

    @Column(name = "recognized_amount", nullable = false, precision = 19, scale = 4)
    var recognizedAmount: BigDecimal = BigDecimal.ZERO,

    @Column(name = "deferred_amount", nullable = false, precision = 19, scale = 4)
    var deferredAmount: BigDecimal,

    @Column(name = "currency", length = 3, nullable = false)
    val currency: String,

    @Column(name = "recognition_start_date", nullable = false)
    val recognitionStartDate: LocalDate,

    @Column(name = "recognition_end_date", nullable = false)
    val recognitionEndDate: LocalDate,

    @Column(name = "status", nullable = false)
    @Enumerated(EnumType.STRING)
    var status: RecognitionStatus = RecognitionStatus.PENDING,

    @Column(name = "created_at", nullable = false, updatable = false)
    val createdAt: Instant = Instant.now(),

    @Column(name = "updated_at", nullable = false)
    var updatedAt: Instant = Instant.now()
) {
    @PreUpdate
    fun onUpdate() {
        updatedAt = Instant.now()
    }
}
