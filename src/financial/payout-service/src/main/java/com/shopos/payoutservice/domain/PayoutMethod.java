package com.shopos.payoutservice.domain;

/**
 * Supported payment rails for vendor payouts.
 */
public enum PayoutMethod {
    /** Standard domestic bank wire/SWIFT transfer. */
    BANK_TRANSFER,

    /** US Automated Clearing House — batch, low-cost. */
    ACH,

    /** International wire transfer (SWIFT/SEPA). */
    WIRE,

    /** PayPal mass-payout API. */
    PAYPAL,

    /** Blockchain / stablecoin transfer. */
    CRYPTO
}
