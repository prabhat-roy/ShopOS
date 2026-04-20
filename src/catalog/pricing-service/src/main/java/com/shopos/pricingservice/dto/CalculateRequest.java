package com.shopos.pricingservice.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;

public record CalculateRequest(

        @NotBlank(message = "productId must not be blank")
        String productId,

        @Min(value = 1, message = "quantity must be at least 1")
        int quantity,

        String segment
) {
    public CalculateRequest {
        if (segment == null || segment.isBlank()) segment = "all";
    }
}
