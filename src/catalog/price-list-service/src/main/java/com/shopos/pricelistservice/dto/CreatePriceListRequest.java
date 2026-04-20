package com.shopos.pricelistservice.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;

public record CreatePriceListRequest(

        @NotBlank(message = "name must not be blank")
        String name,

        @NotBlank(message = "code must not be blank")
        @Pattern(regexp = "^[A-Z0-9_-]+$", message = "code must contain only uppercase letters, digits, hyphens, or underscores")
        String code,

        @NotBlank(message = "currency must not be blank")
        String currency,

        String description
) {
    public CreatePriceListRequest {
        if (currency == null || currency.isBlank()) currency = "USD";
        if (description == null) description = "";
    }
}
