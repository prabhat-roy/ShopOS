package com.shopos.chargeback.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.math.BigDecimal;
import java.time.Instant;

@Data
public class CreateChargebackRequest {

    @NotBlank
    private String paymentId;

    @NotBlank
    private String orderId;

    @NotBlank
    private String customerId;

    @NotNull
    @Positive
    private BigDecimal amount;

    @NotBlank
    private String currency;

    private String reasonCode;
    private String reasonDescription;
    private Instant evidenceDueDate;
}
