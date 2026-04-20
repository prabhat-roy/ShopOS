package com.shopos.invoiceservice.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.UUID;

/**
 * Request payload for creating a new invoice.
 * Invoice number is generated server-side; status defaults to DRAFT.
 */
public record CreateInvoiceRequest(

        @NotNull(message = "orderId is required")
        UUID orderId,

        @NotNull(message = "customerId is required")
        UUID customerId,

        @NotNull(message = "subtotal is required")
        @DecimalMin(value = "0.0", inclusive = true, message = "subtotal must be >= 0")
        BigDecimal subtotal,

        @NotNull(message = "taxAmount is required")
        @DecimalMin(value = "0.0", inclusive = true, message = "taxAmount must be >= 0")
        BigDecimal taxAmount,

        @NotNull(message = "totalAmount is required")
        @DecimalMin(value = "0.01", message = "totalAmount must be > 0")
        BigDecimal totalAmount,

        @Pattern(regexp = "^[A-Z]{3}$", message = "currency must be a 3-letter ISO 4217 code")
        String currency,

        @NotNull(message = "lineItems JSON is required")
        @NotBlank(message = "lineItems JSON must not be blank")
        String lineItems,

        @Size(max = 1000, message = "billingAddress must be <= 1000 characters")
        String billingAddress,

        @NotNull(message = "dueDate is required")
        LocalDate dueDate,

        @Size(max = 2000, message = "notes must be <= 2000 characters")
        String notes
) {}
