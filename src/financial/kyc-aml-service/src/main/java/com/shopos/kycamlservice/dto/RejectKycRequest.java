package com.shopos.kycamlservice.dto;

import jakarta.validation.constraints.NotBlank;

public record RejectKycRequest(

    @NotBlank(message = "reason is required")
    String reason
) {}
