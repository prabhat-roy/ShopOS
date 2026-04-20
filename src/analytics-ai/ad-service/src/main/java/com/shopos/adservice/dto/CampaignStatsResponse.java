package com.shopos.adservice.dto;

import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.util.UUID;

@Data
@Builder
public class CampaignStatsResponse {
    private UUID campaignId;
    private long impressions;
    private long clicks;
    private double ctr;
    private BigDecimal spent;
    private BigDecimal budget;
    private double budgetUtilizationPct;
}
