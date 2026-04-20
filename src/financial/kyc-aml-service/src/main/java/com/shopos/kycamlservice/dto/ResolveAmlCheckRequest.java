package com.shopos.kycamlservice.dto;

import jakarta.validation.constraints.NotBlank;

public record ResolveAmlCheckRequest(

    @NotBlank(message = "resolution is required")
    String resolution,

    @NotBlank(message = "resolvedBy is required")
    String resolvedBy
) {}
