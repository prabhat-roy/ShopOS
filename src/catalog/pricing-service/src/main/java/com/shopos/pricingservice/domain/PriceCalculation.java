package com.shopos.pricingservice.domain;

import java.math.BigDecimal;

public record PriceCalculation(
        String productId,
        int quantity,
        String segment,
        BigDecimal basePrice,
        BigDecimal finalPrice,
        BigDecimal discountPercent,
        String currency
) {}
