package com.shopos.promotionsservice.dto;

import com.shopos.promotionsservice.domain.PromoType;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;
import java.time.Instant;

public record CreatePromotionRequest(

        @NotBlank(message = "code is required")
        String code,

        @NotBlank(message = "name is required")
        String name,

        @NotNull(message = "type is required")
        PromoType type,

        /** Absolute amount (FIXED_AMOUNT type). May be null for other types. */
        BigDecimal discountValue,

        /** Percentage 0–100 (PERCENTAGE type). May be null for other types. */
        BigDecimal discountPercent,

        /** Minimum order subtotal. Defaults to 0 when null. */
        BigDecimal minOrderAmount,

        /** Maximum redemptions. 0 = unlimited. */
        Integer maxUses,

        Instant startsAt,

        Instant expiresAt
) {}
