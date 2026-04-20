package com.shopos.adservice.domain;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "ad_campaigns")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class AdCampaign {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "name", nullable = false, length = 255)
    private String name;

    @Column(name = "advertiser_id", nullable = false)
    private UUID advertiserId;

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false, length = 20)
    @Builder.Default
    private AdStatus status = AdStatus.DRAFT;

    @Enumerated(EnumType.STRING)
    @Column(name = "ad_type", nullable = false, length = 30)
    private AdType adType;

    /**
     * Comma-separated list of category IDs this campaign targets.
     * Stored as plain TEXT to avoid over-normalisation for a simple reference implementation.
     */
    @Column(name = "target_categories", columnDefinition = "TEXT")
    private String targetCategories;

    /**
     * JSON blob describing audience targeting criteria.
     */
    @Column(name = "target_audience", columnDefinition = "TEXT")
    private String targetAudience;

    @Column(name = "budget", precision = 19, scale = 4, nullable = false)
    private BigDecimal budget;

    @Column(name = "spent", precision = 19, scale = 4, nullable = false)
    @Builder.Default
    private BigDecimal spent = BigDecimal.ZERO;

    @Column(name = "impressions", nullable = false)
    @Builder.Default
    private long impressions = 0L;

    @Column(name = "clicks", nullable = false)
    @Builder.Default
    private long clicks = 0L;

    @Column(name = "start_date", nullable = false)
    private LocalDate startDate;

    @Column(name = "end_date", nullable = false)
    private LocalDate endDate;

    @Column(name = "image_url", length = 2048)
    private String imageUrl;

    @Column(name = "target_url", length = 2048)
    private String targetUrl;

    @Column(name = "bid_amount", precision = 19, scale = 4)
    private BigDecimal bidAmount;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;
}
