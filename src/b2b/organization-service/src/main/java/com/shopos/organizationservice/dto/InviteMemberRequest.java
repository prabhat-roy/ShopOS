package com.shopos.organizationservice.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.util.UUID;

public record InviteMemberRequest(
        @NotNull(message = "userId is required")
        UUID userId,

        @NotBlank(message = "role is required")
        String role,

        String department,
        String jobTitle
) {
}
