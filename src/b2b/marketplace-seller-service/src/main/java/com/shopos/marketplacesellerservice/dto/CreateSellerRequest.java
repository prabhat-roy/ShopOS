package com.shopos.marketplacesellerservice.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.DecimalMax;
import jakarta.validation.constraints.DecimalMin;

import java.math.BigDecimal;
import java.util.UUID;

/**
 * Request body for onboarding a new marketplace seller.
 */
public record CreateSellerRequest(

        @NotNull(message = "Organisation ID is required")
        UUID orgId,

        @NotBlank(message = "Display name is required")
        String displayName,

        String description,

        @DecimalMin(value = "0.00", message = "Commission rate must be at least 0%")
        @DecimalMax(value = "100.00", message = "Commission rate cannot exceed 100%")
        BigDecimal commissionRate
) {}
