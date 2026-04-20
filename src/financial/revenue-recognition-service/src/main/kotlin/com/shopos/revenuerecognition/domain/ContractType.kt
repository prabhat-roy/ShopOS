package com.shopos.revenuerecognition.domain

enum class ContractType {
    /** Point-in-time recognition — e.g., one-time product sale */
    ONE_TIME,
    /** Straight-line recognition over subscription period */
    SUBSCRIPTION,
    /** Gift card — recognized upon redemption */
    GIFT_CARD,
    /** Multi-element arrangement — allocated across performance obligations */
    MULTI_ELEMENT
}
