package com.shopos.adservice.dto;

import com.shopos.adservice.domain.AdType;
import jakarta.validation.constraints.*;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.UUID;

@Data
public class CreateCampaignRequest {

    @NotBlank(message = "Campaign name is required")
    @Size(min = 1, max = 255, message = "Campaign name must be between 1 and 255 characters")
    private String name;

    @NotNull(message = "Advertiser ID is required")
    private UUID advertiserId;

    @NotNull(message = "Ad type is required")
    private AdType adType;

    private String targetCategories;

    private String targetAudience;

    @NotNull(message = "Budget is required")
    @DecimalMin(value = "0.01", message = "Budget must be greater than 0")
    private BigDecimal budget;

    @NotNull(message = "Start date is required")
    private LocalDate startDate;

    @NotNull(message = "End date is required")
    private LocalDate endDate;

    @Size(max = 2048, message = "Image URL must not exceed 2048 characters")
    private String imageUrl;

    @NotBlank(message = "Target URL is required")
    @Size(max = 2048, message = "Target URL must not exceed 2048 characters")
    private String targetUrl;

    @DecimalMin(value = "0.01", message = "Bid amount must be greater than 0")
    private BigDecimal bidAmount;
}
