package com.shopos.pricingservice.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;
import java.time.OffsetDateTime;

public record SetPriceRequest(

        @NotBlank(message = "productId must not be blank")
        String productId,

        @NotBlank(message = "currency must not be blank")
        String currency,

        @NotNull(message = "basePrice is required")
        @DecimalMin(value = "0.0", inclusive = false, message = "basePrice must be greater than zero")
        BigDecimal basePrice,

        BigDecimal salePrice,

        @Min(value = 1, message = "minQty must be at least 1")
        int minQty,

        String segment,

        OffsetDateTime startAt,

        OffsetDateTime endAt
) {
    public SetPriceRequest {
        if (currency == null || currency.isBlank()) currency = "USD";
        if (segment == null || segment.isBlank()) segment = "all";
        if (minQty < 1) minQty = 1;
    }
}
