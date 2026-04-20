package com.shopos.supplierportalservice.dto;

import jakarta.validation.constraints.*;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.UUID;

@Data
public class CreateInvoiceRequest {

    @NotNull(message = "vendorId is required")
    private UUID vendorId;

    private UUID purchaseOrderId;

    @NotBlank(message = "invoiceNumber is required")
    @Size(max = 100, message = "invoiceNumber must be at most 100 characters")
    private String invoiceNumber;

    @NotNull(message = "amount is required")
    @DecimalMin(value = "0.0001", message = "amount must be positive")
    @Digits(integer = 15, fraction = 4, message = "amount format invalid")
    private BigDecimal amount;

    @Size(min = 3, max = 3, message = "currency must be a 3-letter ISO code")
    private String currency = "USD";

    private LocalDate dueDate;

    private String lineItems;

    private String notes;
}
