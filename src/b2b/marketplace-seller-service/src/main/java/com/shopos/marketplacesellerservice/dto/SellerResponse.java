package com.shopos.marketplacesellerservice.dto;

import com.shopos.marketplacesellerservice.domain.Seller;
import com.shopos.marketplacesellerservice.domain.SellerStatus;
import com.shopos.marketplacesellerservice.domain.SellerTier;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.UUID;

/**
 * Read-model response object for a marketplace seller.
 */
public record SellerResponse(
        UUID id,
        UUID orgId,
        String displayName,
        String description,
        SellerStatus status,
        SellerTier tier,
        BigDecimal commissionRate,
        BigDecimal rating,
        BigDecimal totalSales,
        int totalOrders,
        int productCount,
        BigDecimal returnRate,
        Instant onboardedAt,
        Instant createdAt,
        Instant updatedAt
) {

    /**
     * Factory method mapping from the JPA entity to this response record.
     */
    public static SellerResponse from(Seller seller) {
        return new SellerResponse(
                seller.getId(),
                seller.getOrgId(),
                seller.getDisplayName(),
                seller.getDescription(),
                seller.getStatus(),
                seller.getTier(),
                seller.getCommissionRate(),
                seller.getRating(),
                seller.getTotalSales(),
                seller.getTotalOrders(),
                seller.getProductCount(),
                seller.getReturnRate(),
                seller.getOnboardedAt(),
                seller.getCreatedAt(),
                seller.getUpdatedAt()
        );
    }
}
