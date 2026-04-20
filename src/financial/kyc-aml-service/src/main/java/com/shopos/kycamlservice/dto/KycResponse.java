package com.shopos.kycamlservice.dto;

import com.shopos.kycamlservice.domain.KycRecord;
import com.shopos.kycamlservice.domain.KycStatus;
import com.shopos.kycamlservice.domain.RiskLevel;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

public record KycResponse(
    UUID id,
    UUID customerId,
    String firstName,
    String lastName,
    LocalDate dateOfBirth,
    String nationality,
    String documentType,
    String documentNumber,
    LocalDate documentExpiry,
    KycStatus status,
    RiskLevel riskLevel,
    LocalDateTime verifiedAt,
    LocalDateTime expiresAt,
    String rejectionReason,
    String notes,
    LocalDateTime createdAt,
    LocalDateTime updatedAt
) {

    public static KycResponse from(KycRecord record) {
        return new KycResponse(
            record.getId(),
            record.getCustomerId(),
            record.getFirstName(),
            record.getLastName(),
            record.getDateOfBirth(),
            record.getNationality(),
            record.getDocumentType(),
            record.getDocumentNumber(),
            record.getDocumentExpiry(),
            record.getStatus(),
            record.getRiskLevel(),
            record.getVerifiedAt(),
            record.getExpiresAt(),
            record.getRejectionReason(),
            record.getNotes(),
            record.getCreatedAt(),
            record.getUpdatedAt()
        );
    }
}
