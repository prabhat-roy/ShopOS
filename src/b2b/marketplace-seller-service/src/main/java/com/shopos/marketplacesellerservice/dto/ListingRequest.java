package com.shopos.marketplacesellerservice.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

/**
 * Request body for creating or updating a seller product listing.
 */
public record ListingRequest(

        @NotBlank(message = "Product ID is required")
        String productId,

        @NotBlank(message = "SKU is required")
        String sku,

        @NotBlank(message = "Seller SKU is required")
        String sellerSku,

        @NotNull(message = "Listing price is required")
        @DecimalMin(value = "0.01", message = "Listing price must be greater than 0")
        BigDecimal listingPrice,

        @Min(value = 0, message = "Stock quantity must be non-negative")
        int stockQuantity
) {}
