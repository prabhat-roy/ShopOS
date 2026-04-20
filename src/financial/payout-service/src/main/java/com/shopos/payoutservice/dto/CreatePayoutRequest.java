package com.shopos.payoutservice.dto;

import com.shopos.payoutservice.domain.PayoutMethod;
import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

/**
 * Request payload for creating a new vendor payout.
 * Reference is generated server-side; status defaults to PENDING.
 */
public record CreatePayoutRequest(

        @NotNull(message = "vendorId is required")
        UUID vendorId,

        @NotNull(message = "amount is required")
        @DecimalMin(value = "0.01", message = "amount must be greater than 0")
        BigDecimal amount,

        @Pattern(regexp = "^[A-Z]{3}$", message = "currency must be a 3-letter ISO 4217 code")
        String currency,

        @NotNull(message = "method is required")
        PayoutMethod method,

        @Size(max = 2000, message = "bankAccount details must be <= 2000 characters")
        String bankAccount,

        /**
         * Optional future datetime at which this payout should be processed.
         * If null, the payout is immediately eligible for processing.
         */
        LocalDateTime scheduledAt
) {}
