package com.shopos.kycamlservice.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Pattern;

import java.util.UUID;

public record RunAmlCheckRequest(

    @NotNull(message = "customerId is required")
    UUID customerId,

    @NotBlank(message = "checkType is required")
    @Pattern(
        regexp = "SANCTIONS|PEP|ADVERSE_MEDIA|TRANSACTION_MONITORING",
        message = "checkType must be one of: SANCTIONS, PEP, ADVERSE_MEDIA, TRANSACTION_MONITORING"
    )
    String checkType
) {}
