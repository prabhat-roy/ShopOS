package com.shopos.kycamlservice.dto;

import com.shopos.kycamlservice.domain.AmlCheck;

import java.time.LocalDateTime;
import java.util.UUID;

public record AmlCheckResponse(
    UUID id,
    UUID customerId,
    String checkType,
    String result,
    int riskScore,
    String details,
    LocalDateTime checkedAt,
    LocalDateTime resolvedAt,
    String resolvedBy,
    String resolution
) {

    public static AmlCheckResponse from(AmlCheck check) {
        return new AmlCheckResponse(
            check.getId(),
            check.getCustomerId(),
            check.getCheckType(),
            check.getResult(),
            check.getRiskScore(),
            check.getDetails(),
            check.getCheckedAt(),
            check.getResolvedAt(),
            check.getResolvedBy(),
            check.getResolution()
        );
    }
}
