package com.shopos.supplierportalservice.dto;

import jakarta.validation.constraints.*;
import lombok.Data;

import java.math.BigDecimal;
import java.util.UUID;

@Data
public class UpsertCatalogItemRequest {

    @NotNull(message = "vendorId is required")
    private UUID vendorId;

    @NotBlank(message = "productId is required")
    @Size(max = 200)
    private String productId;

    @NotBlank(message = "sku is required")
    @Size(max = 100)
    private String sku;

    @NotBlank(message = "productName is required")
    @Size(max = 500)
    private String productName;

    @NotNull(message = "unitPrice is required")
    @DecimalMin(value = "0.0001", message = "unitPrice must be positive")
    @Digits(integer = 15, fraction = 4)
    private BigDecimal unitPrice;

    @Size(min = 3, max = 3, message = "currency must be a 3-letter ISO code")
    private String currency = "USD";

    @Min(value = 1, message = "minOrderQty must be at least 1")
    private int minOrderQty = 1;

    @Min(value = 0, message = "leadTimeDays must be non-negative")
    private int leadTimeDays;
}
