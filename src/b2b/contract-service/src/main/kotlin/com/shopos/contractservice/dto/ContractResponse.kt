package com.shopos.contractservice.dto

import com.shopos.contractservice.domain.Contract
import com.shopos.contractservice.domain.ContractStatus
import com.shopos.contractservice.domain.ContractType
import java.math.BigDecimal
import java.time.LocalDate
import java.time.LocalDateTime
import java.util.UUID

data class ContractResponse(
    val id: UUID?,
    val orgId: UUID,
    val vendorId: UUID?,
    val title: String,
    val type: ContractType,
    val status: ContractStatus,
    val description: String?,
    val terms: String?,
    val value: BigDecimal?,
    val currency: String,
    val startDate: LocalDate,
    val endDate: LocalDate,
    val autoRenew: Boolean,
    val signedByBuyer: Boolean,
    val signedByVendor: Boolean,
    val signedAt: LocalDateTime?,
    val terminationReason: String?,
    val createdBy: String,
    val createdAt: LocalDateTime?,
    val updatedAt: LocalDateTime?
) {
    companion object {
        fun from(c: Contract): ContractResponse = ContractResponse(
            id = c.id,
            orgId = c.orgId,
            vendorId = c.vendorId,
            title = c.title,
            type = c.type,
            status = c.status,
            description = c.description,
            terms = c.terms,
            value = c.value,
            currency = c.currency,
            startDate = c.startDate,
            endDate = c.endDate,
            autoRenew = c.autoRenew,
            signedByBuyer = c.signedByBuyer,
            signedByVendor = c.signedByVendor,
            signedAt = c.signedAt,
            terminationReason = c.terminationReason,
            createdBy = c.createdBy,
            createdAt = c.createdAt,
            updatedAt = c.updatedAt
        )
    }
}
