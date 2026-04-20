package com.shopos.adservice.domain;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "ad_impressions", indexes = {
        @Index(name = "idx_impression_campaign_id", columnList = "campaign_id"),
        @Index(name = "idx_impression_session_id",  columnList = "session_id")
})
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class AdImpression {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "campaign_id", nullable = false)
    private UUID campaignId;

    @Column(name = "user_id", length = 255)
    private String userId;

    @Column(name = "session_id", nullable = false, length = 255)
    private String sessionId;

    @Column(name = "placement_id", nullable = false, length = 255)
    private String placementId;

    @CreationTimestamp
    @Column(name = "shown_at", nullable = false, updatable = false)
    private LocalDateTime shownAt;

    @Column(name = "clicked", nullable = false)
    @Builder.Default
    private boolean clicked = false;

    @Column(name = "clicked_at")
    private LocalDateTime clickedAt;
}
