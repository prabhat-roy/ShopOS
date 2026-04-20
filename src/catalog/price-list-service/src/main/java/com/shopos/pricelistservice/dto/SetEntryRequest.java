package com.shopos.pricelistservice.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record SetEntryRequest(

        @NotBlank(message = "productId must not be blank")
        String productId,

        @NotNull(message = "price is required")
        @DecimalMin(value = "0.0", inclusive = false, message = "price must be greater than zero")
        BigDecimal price
) {}
