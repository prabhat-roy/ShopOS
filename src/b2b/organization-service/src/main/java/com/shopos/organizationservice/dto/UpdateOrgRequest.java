package com.shopos.organizationservice.dto;

import com.shopos.organizationservice.domain.OrgType;

import java.math.BigDecimal;
import java.util.UUID;

public record UpdateOrgRequest(
        String name,
        String phone,
        String website,
        OrgType type,
        String industry,
        String taxId,
        String country,
        String address,
        Integer employeeCount,
        BigDecimal creditLimit,
        UUID parentOrgId,
        String settings
) {
}
