package com.shopos.paymentservice.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record RefundRequest(

        @NotNull(message = "amount is required")
        @DecimalMin(value = "0.01", message = "refund amount must be greater than zero")
        BigDecimal amount,

        @NotBlank(message = "reason is required")
        String reason
) {}
