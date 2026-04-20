package com.shopos.pricelistservice.dto;

/**
 * Partial update payload for a price list. All fields are optional — null means "no change".
 */
public record UpdatePriceListRequest(
        String name,
        String description,
        Boolean active
) {}
