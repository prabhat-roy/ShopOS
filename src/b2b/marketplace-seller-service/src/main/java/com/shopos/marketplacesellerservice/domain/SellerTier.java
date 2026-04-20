package com.shopos.marketplacesellerservice.domain;

/**
 * Performance-based tier that determines commission rate, visibility and perks.
 *
 * <ul>
 *   <li>BRONZE  — fewer than 50 fulfilled orders</li>
 *   <li>SILVER  — 50–199 fulfilled orders</li>
 *   <li>GOLD    — 200–999 fulfilled orders</li>
 *   <li>PLATINUM — 1000+ fulfilled orders</li>
 * </ul>
 */
public enum SellerTier {
    BRONZE,
    SILVER,
    GOLD,
    PLATINUM
}
