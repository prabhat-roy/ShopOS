package com.shopos.paymentservice.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record CreatePaymentRequest(

        @NotBlank(message = "orderId is required")
        String orderId,

        @NotBlank(message = "customerId is required")
        String customerId,

        @NotNull(message = "amount is required")
        @DecimalMin(value = "0.01", message = "amount must be greater than zero")
        BigDecimal amount,

        @NotBlank(message = "currency is required")
        String currency,

        @NotBlank(message = "provider is required")
        String provider,

        @NotBlank(message = "paymentMethodToken is required")
        String paymentMethodToken
) {}
