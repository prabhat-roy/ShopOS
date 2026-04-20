package com.shopos.organizationservice.dto;

import com.shopos.organizationservice.domain.OrgType;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

public record CreateOrgRequest(
        @NotBlank(message = "Organization name is required")
        String name,

        @NotBlank(message = "Email is required")
        @Email(message = "Email must be valid")
        String email,

        String phone,
        String website,
        OrgType type,
        String industry,
        String taxId,
        String country,
        String address,
        Integer employeeCount
) {
}
