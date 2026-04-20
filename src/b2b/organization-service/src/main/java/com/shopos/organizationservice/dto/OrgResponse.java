package com.shopos.organizationservice.dto;

import com.shopos.organizationservice.domain.OrgStatus;
import com.shopos.organizationservice.domain.OrgType;
import com.shopos.organizationservice.domain.Organization;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

public record OrgResponse(
        UUID id,
        String name,
        String slug,
        String email,
        String phone,
        String website,
        OrgType type,
        OrgStatus status,
        String industry,
        String taxId,
        String country,
        String address,
        int employeeCount,
        BigDecimal creditLimit,
        UUID parentOrgId,
        LocalDateTime createdAt,
        LocalDateTime updatedAt
) {
    public static OrgResponse from(Organization o) {
        return new OrgResponse(
                o.getId(),
                o.getName(),
                o.getSlug(),
                o.getEmail(),
                o.getPhone(),
                o.getWebsite(),
                o.getType(),
                o.getStatus(),
                o.getIndustry(),
                o.getTaxId(),
                o.getCountry(),
                o.getAddress(),
                o.getEmployeeCount(),
                o.getCreditLimit(),
                o.getParentOrgId(),
                o.getCreatedAt(),
                o.getUpdatedAt()
        );
    }
}
