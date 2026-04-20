package com.shopos.promotionsservice.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record ValidateRequest(

        @NotBlank(message = "code is required")
        String code,

        @NotNull(message = "orderAmount is required")
        @DecimalMin(value = "0.00", message = "orderAmount must be non-negative")
        BigDecimal orderAmount,

        @NotBlank(message = "customerId is required")
        String customerId
) {}
