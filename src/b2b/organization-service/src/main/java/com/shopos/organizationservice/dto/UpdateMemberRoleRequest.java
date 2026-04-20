package com.shopos.organizationservice.dto;

import jakarta.validation.constraints.NotBlank;

public record UpdateMemberRoleRequest(
        @NotBlank(message = "role is required")
        String role
) {
}
