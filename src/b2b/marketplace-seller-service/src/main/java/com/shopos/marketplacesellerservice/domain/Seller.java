package com.shopos.marketplacesellerservice.domain;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.UUID;

/**
 * JPA entity representing a marketplace seller account.
 */
@Entity
@Table(name = "sellers", indexes = {
        @Index(name = "idx_sellers_org_id", columnList = "org_id", unique = true),
        @Index(name = "idx_sellers_status", columnList = "status"),
        @Index(name = "idx_sellers_tier", columnList = "tier")
})
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Seller {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(nullable = false, updatable = false)
    private UUID id;

    /**
     * Foreign key to the organisation-service — uniquely identifies the legal entity.
     */
    @Column(name = "org_id", nullable = false, unique = true)
    private UUID orgId;

    @Column(name = "display_name", nullable = false, length = 255)
    private String displayName;

    @Column(name = "description", columnDefinition = "TEXT")
    private String description;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private SellerStatus status;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    @Builder.Default
    private SellerTier tier = SellerTier.BRONZE;

    /**
     * Commission rate as a percentage (e.g. 15.00 = 15 %).
     */
    @Column(name = "commission_rate", precision = 5, scale = 2)
    @Builder.Default
    private BigDecimal commissionRate = new BigDecimal("15.00");

    /**
     * Aggregate seller rating 0–5 (updated by review domain events).
     */
    @Column(name = "rating", precision = 3, scale = 2)
    @Builder.Default
    private BigDecimal rating = BigDecimal.ZERO;

    @Column(name = "total_sales", precision = 19, scale = 2)
    @Builder.Default
    private BigDecimal totalSales = BigDecimal.ZERO;

    @Column(name = "total_orders")
    @Builder.Default
    private int totalOrders = 0;

    @Column(name = "product_count")
    @Builder.Default
    private int productCount = 0;

    /**
     * Return rate as a percentage (e.g. 3.50 = 3.5 %).
     */
    @Column(name = "return_rate", precision = 5, scale = 2)
    @Builder.Default
    private BigDecimal returnRate = BigDecimal.ZERO;

    @Column(name = "onboarded_at")
    private Instant onboardedAt;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;
}
