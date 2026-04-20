package com.shopos.marketplacesellerservice.domain;

/**
 * Lifecycle states of a marketplace seller account.
 */
public enum SellerStatus {
    /**
     * Seller has submitted an onboarding application; awaiting review.
     */
    PENDING,

    /**
     * Seller has been approved and can list products and receive orders.
     */
    ACTIVE,

    /**
     * Seller has been temporarily suspended — listings hidden, orders paused.
     */
    SUSPENDED,

    /**
     * Seller account has been permanently closed; cannot be reactivated.
     */
    TERMINATED
}
