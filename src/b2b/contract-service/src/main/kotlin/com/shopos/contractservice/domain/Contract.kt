package com.shopos.contractservice.domain

import jakarta.persistence.*
import org.hibernate.annotations.CreationTimestamp
import org.hibernate.annotations.UpdateTimestamp
import java.math.BigDecimal
import java.time.LocalDate
import java.time.LocalDateTime
import java.util.UUID

@Entity
@Table(name = "contracts")
data class Contract(

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(updatable = false, nullable = false)
    val id: UUID? = null,

    @Column(name = "org_id", nullable = false)
    val orgId: UUID,

    @Column(name = "vendor_id")
    var vendorId: UUID? = null,

    @Column(nullable = false)
    var title: String,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    var type: ContractType,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    var status: ContractStatus = ContractStatus.DRAFT,

    @Column(columnDefinition = "TEXT")
    var description: String? = null,

    @Column(columnDefinition = "TEXT")
    var terms: String? = null,

    @Column(precision = 19, scale = 4)
    var value: BigDecimal? = null,

    @Column(length = 3)
    var currency: String = "USD",

    @Column(name = "start_date", nullable = false)
    var startDate: LocalDate,

    @Column(name = "end_date", nullable = false)
    var endDate: LocalDate,

    @Column(name = "auto_renew")
    var autoRenew: Boolean = false,

    @Column(name = "signed_by_buyer")
    var signedByBuyer: Boolean = false,

    @Column(name = "signed_by_vendor")
    var signedByVendor: Boolean = false,

    @Column(name = "signed_at")
    var signedAt: LocalDateTime? = null,

    @Column(name = "termination_reason", columnDefinition = "TEXT")
    var terminationReason: String? = null,

    @Column(name = "created_by", nullable = false)
    val createdBy: String,

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    val createdAt: LocalDateTime? = null,

    @UpdateTimestamp
    @Column(name = "updated_at")
    var updatedAt: LocalDateTime? = null
)
