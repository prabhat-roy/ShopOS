package com.shopos.chargeback.dto;

import com.shopos.chargeback.domain.ChargebackStatus;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.UUID;

@Data
@Builder
public class ChargebackResponse {
    private UUID id;
    private String paymentId;
    private String orderId;
    private String customerId;
    private BigDecimal amount;
    private String currency;
    private ChargebackStatus status;
    private String reasonCode;
    private String reasonDescription;
    private Instant evidenceDueDate;
    private Instant evidenceSubmittedAt;
    private Instant resolvedAt;
    private Instant createdAt;
    private Instant updatedAt;
}
