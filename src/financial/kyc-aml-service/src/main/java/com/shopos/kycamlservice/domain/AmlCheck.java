package com.shopos.kycamlservice.domain;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(
    name = "aml_checks",
    indexes = {
        @Index(name = "idx_aml_customer_id", columnList = "customer_id"),
        @Index(name = "idx_aml_check_type_result", columnList = "check_type, result")
    }
)
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AmlCheck {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "customer_id", nullable = false)
    private UUID customerId;

    /**
     * One of: SANCTIONS, PEP, ADVERSE_MEDIA, TRANSACTION_MONITORING
     */
    @Column(name = "check_type", nullable = false, length = 40)
    private String checkType;

    /**
     * One of: CLEAR, FLAGGED, REVIEW_REQUIRED
     */
    @Column(name = "result", nullable = false, length = 20)
    private String result;

    @Column(name = "risk_score", nullable = false)
    private int riskScore;

    @Column(name = "details", columnDefinition = "TEXT")
    private String details;

    @Column(name = "checked_at", nullable = false)
    private LocalDateTime checkedAt;

    @Column(name = "resolved_at")
    private LocalDateTime resolvedAt;

    @Column(name = "resolved_by", length = 200)
    private String resolvedBy;

    @Column(name = "resolution", columnDefinition = "TEXT")
    private String resolution;
}
