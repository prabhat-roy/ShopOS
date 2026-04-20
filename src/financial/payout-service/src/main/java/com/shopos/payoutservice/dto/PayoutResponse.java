package com.shopos.payoutservice.dto;

import com.shopos.payoutservice.domain.Payout;
import com.shopos.payoutservice.domain.PayoutMethod;
import com.shopos.payoutservice.domain.PayoutStatus;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

/**
 * Outbound representation of a Payout resource.
 */
public record PayoutResponse(
        UUID id,
        UUID vendorId,
        BigDecimal amount,
        String currency,
        PayoutStatus status,
        PayoutMethod method,
        String reference,
        String bankAccount,
        String failureReason,
        LocalDateTime scheduledAt,
        LocalDateTime processedAt,
        LocalDateTime createdAt,
        LocalDateTime updatedAt
) {

    /**
     * Factory method that maps a {@link Payout} entity to a {@link PayoutResponse} DTO.
     *
     * @param payout the entity to convert
     * @return an immutable response record
     */
    public static PayoutResponse from(Payout payout) {
        return new PayoutResponse(
                payout.getId(),
                payout.getVendorId(),
                payout.getAmount(),
                payout.getCurrency(),
                payout.getStatus(),
                payout.getMethod(),
                payout.getReference(),
                payout.getBankAccount(),
                payout.getFailureReason(),
                payout.getScheduledAt(),
                payout.getProcessedAt(),
                payout.getCreatedAt(),
                payout.getUpdatedAt()
        );
    }
}
