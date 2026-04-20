package com.shopos.chargeback.service;

import com.shopos.chargeback.domain.Chargeback;
import com.shopos.chargeback.domain.ChargebackStatus;
import com.shopos.chargeback.dto.ChargebackResponse;
import com.shopos.chargeback.dto.CreateChargebackRequest;
import com.shopos.chargeback.repository.ChargebackRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class ChargebackService {

    private final ChargebackRepository chargebackRepository;

    @Transactional
    public ChargebackResponse createChargeback(CreateChargebackRequest request) {
        Chargeback chargeback = Chargeback.builder()
                .paymentId(request.getPaymentId())
                .orderId(request.getOrderId())
                .customerId(request.getCustomerId())
                .amount(request.getAmount())
                .currency(request.getCurrency())
                .status(ChargebackStatus.OPEN)
                .reasonCode(request.getReasonCode())
                .reasonDescription(request.getReasonDescription())
                .evidenceDueDate(request.getEvidenceDueDate())
                .build();

        Chargeback saved = chargebackRepository.save(chargeback);
        log.info("Chargeback created id={} paymentId={}", saved.getId(), saved.getPaymentId());
        return toResponse(saved);
    }

    @Transactional(readOnly = true)
    public ChargebackResponse getChargeback(UUID id) {
        Chargeback chargeback = chargebackRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Chargeback not found: " + id));
        return toResponse(chargeback);
    }

    @Transactional(readOnly = true)
    public List<ChargebackResponse> getChargebacksByCustomer(String customerId) {
        return chargebackRepository.findByCustomerId(customerId)
                .stream().map(this::toResponse).collect(Collectors.toList());
    }

    @Transactional
    public ChargebackResponse submitEvidence(UUID id) {
        Chargeback chargeback = chargebackRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Chargeback not found: " + id));
        chargeback.setStatus(ChargebackStatus.EVIDENCE_SUBMITTED);
        chargeback.setEvidenceSubmittedAt(Instant.now());
        Chargeback saved = chargebackRepository.save(chargeback);
        log.info("Evidence submitted for chargeback id={}", id);
        return toResponse(saved);
    }

    @Transactional
    public ChargebackResponse resolveChargeback(UUID id, ChargebackStatus resolution) {
        if (resolution != ChargebackStatus.WON && resolution != ChargebackStatus.LOST) {
            throw new IllegalArgumentException("Resolution must be WON or LOST");
        }
        Chargeback chargeback = chargebackRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Chargeback not found: " + id));
        chargeback.setStatus(resolution);
        chargeback.setResolvedAt(Instant.now());
        Chargeback saved = chargebackRepository.save(chargeback);
        log.info("Chargeback resolved id={} status={}", id, resolution);
        return toResponse(saved);
    }

    private ChargebackResponse toResponse(Chargeback c) {
        return ChargebackResponse.builder()
                .id(c.getId())
                .paymentId(c.getPaymentId())
                .orderId(c.getOrderId())
                .customerId(c.getCustomerId())
                .amount(c.getAmount())
                .currency(c.getCurrency())
                .status(c.getStatus())
                .reasonCode(c.getReasonCode())
                .reasonDescription(c.getReasonDescription())
                .evidenceDueDate(c.getEvidenceDueDate())
                .evidenceSubmittedAt(c.getEvidenceSubmittedAt())
                .resolvedAt(c.getResolvedAt())
                .createdAt(c.getCreatedAt())
                .updatedAt(c.getUpdatedAt())
                .build();
    }
}
