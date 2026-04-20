package com.shopos.adservice.dto;

import com.shopos.adservice.domain.AdCampaign;
import com.shopos.adservice.domain.AdStatus;
import com.shopos.adservice.domain.AdType;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
public class CampaignResponse {

    private UUID id;
    private String name;
    private UUID advertiserId;
    private AdStatus status;
    private AdType adType;
    private String targetCategories;
    private String targetAudience;
    private BigDecimal budget;
    private BigDecimal spent;
    private long impressions;
    private long clicks;
    private double ctr;
    private LocalDate startDate;
    private LocalDate endDate;
    private String imageUrl;
    private String targetUrl;
    private BigDecimal bidAmount;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    public static CampaignResponse from(AdCampaign campaign) {
        double ctr = campaign.getImpressions() == 0
                ? 0.0
                : BigDecimal.valueOf((double) campaign.getClicks() / campaign.getImpressions())
                            .setScale(4, RoundingMode.HALF_UP)
                            .doubleValue();

        return CampaignResponse.builder()
                .id(campaign.getId())
                .name(campaign.getName())
                .advertiserId(campaign.getAdvertiserId())
                .status(campaign.getStatus())
                .adType(campaign.getAdType())
                .targetCategories(campaign.getTargetCategories())
                .targetAudience(campaign.getTargetAudience())
                .budget(campaign.getBudget())
                .spent(campaign.getSpent())
                .impressions(campaign.getImpressions())
                .clicks(campaign.getClicks())
                .ctr(ctr)
                .startDate(campaign.getStartDate())
                .endDate(campaign.getEndDate())
                .imageUrl(campaign.getImageUrl())
                .targetUrl(campaign.getTargetUrl())
                .bidAmount(campaign.getBidAmount())
                .createdAt(campaign.getCreatedAt())
                .updatedAt(campaign.getUpdatedAt())
                .build();
    }
}
