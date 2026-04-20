package com.shopos.purchaseorderservice.domain

enum class POStatus {
    DRAFT,
    SUBMITTED,
    APPROVED,
    REJECTED,
    PARTIALLY_RECEIVED,
    FULLY_RECEIVED,
    CANCELLED
}
