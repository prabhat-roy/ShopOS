package com.shopos.organizationservice.domain;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "organizations")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Organization {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(updatable = false, nullable = false)
    private UUID id;

    @Column(nullable = false)
    private String name;

    @Column(unique = true, nullable = false)
    private String slug;

    @Column(unique = true, nullable = false)
    private String email;

    private String phone;

    private String website;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private OrgType type;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    @Builder.Default
    private OrgStatus status = OrgStatus.PENDING_VERIFICATION;

    private String industry;

    @Column(name = "tax_id")
    private String taxId;

    private String country;

    @Column(columnDefinition = "TEXT")
    private String address;

    @Column(name = "employee_count")
    @Builder.Default
    private int employeeCount = 0;

    @Column(name = "credit_limit", precision = 19, scale = 4)
    @Builder.Default
    private BigDecimal creditLimit = BigDecimal.ZERO;

    @Column(name = "parent_org_id")
    private UUID parentOrgId;

    @Column(columnDefinition = "TEXT")
    private String settings;

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;
}
