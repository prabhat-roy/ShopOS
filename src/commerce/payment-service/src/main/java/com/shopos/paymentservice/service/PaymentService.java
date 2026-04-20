package com.shopos.paymentservice.service;

import com.shopos.paymentservice.domain.Payment;
import com.shopos.paymentservice.domain.PaymentStatus;
import com.shopos.paymentservice.dto.CreatePaymentRequest;
import com.shopos.paymentservice.dto.RefundRequest;
import com.shopos.paymentservice.event.PaymentEventPublisher;
import com.shopos.paymentservice.repository.PaymentRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.NoSuchElementException;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class PaymentService {

    private final PaymentRepository paymentRepository;
    private final PaymentEventPublisher eventPublisher;

    /**
     * Creates a payment record in PENDING state, immediately simulates a
     * provider authorisation and transitions to AUTHORIZED, then publishes
     * the commerce.payment.processed event.
     */
    @Transactional
    public Payment createPayment(CreatePaymentRequest req) {
        Payment payment = Payment.builder()
                .orderId(req.orderId())
                .customerId(req.customerId())
                .amount(req.amount())
                .currency(req.currency())
                .provider(req.provider())
                .status(PaymentStatus.PENDING)
                .build();

        payment = paymentRepository.save(payment);
        log.info("Payment created id={} orderId={}", payment.getId(), payment.getOrderId());

        // Simulate provider authorisation (no real gateway call in skeleton)
        payment.setStatus(PaymentStatus.AUTHORIZED);
        payment.setProviderTxId("sim_" + UUID.randomUUID().toString().replace("-", "").substring(0, 16));
        payment = paymentRepository.save(payment);
        log.info("Payment authorised id={}", payment.getId());

        eventPublisher.publishProcessed(payment);
        return payment;
    }

    /**
     * Transitions AUTHORIZED → CAPTURED.
     */
    @Transactional
    public Payment capturePayment(UUID id) {
        Payment payment = findOrThrow(id);

        if (payment.getStatus() != PaymentStatus.AUTHORIZED) {
            throw new IllegalStateException(
                    "Payment " + id + " cannot be captured from status " + payment.getStatus());
        }

        payment.setStatus(PaymentStatus.CAPTURED);
        payment = paymentRepository.save(payment);
        log.info("Payment captured id={}", id);
        return payment;
    }

    /**
     * Refunds a CAPTURED payment.
     * If the refund amount exceeds the original amount the status is set to
     * FAILED and a commerce.payment.failed event is published instead.
     */
    @Transactional
    public Payment refundPayment(UUID id, RefundRequest req) {
        Payment payment = findOrThrow(id);

        if (payment.getStatus() != PaymentStatus.CAPTURED) {
            throw new IllegalStateException(
                    "Payment " + id + " cannot be refunded from status " + payment.getStatus());
        }

        if (req.amount().compareTo(payment.getAmount()) > 0) {
            payment.setStatus(PaymentStatus.FAILED);
            payment = paymentRepository.save(payment);
            eventPublisher.publishFailed(payment,
                    "Refund amount " + req.amount() + " exceeds original " + payment.getAmount());
            log.warn("Refund rejected — amount exceeds original for paymentId={}", id);
            return payment;
        }

        payment.setStatus(PaymentStatus.REFUNDED);
        payment = paymentRepository.save(payment);
        log.info("Payment refunded id={} amount={}", id, req.amount());
        eventPublisher.publishProcessed(payment);
        return payment;
    }

    /**
     * Cancels a PENDING or AUTHORIZED payment.
     */
    @Transactional
    public Payment cancelPayment(UUID id) {
        Payment payment = findOrThrow(id);

        if (payment.getStatus() != PaymentStatus.PENDING
                && payment.getStatus() != PaymentStatus.AUTHORIZED) {
            throw new IllegalStateException(
                    "Payment " + id + " cannot be cancelled from status " + payment.getStatus());
        }

        payment.setStatus(PaymentStatus.CANCELLED);
        payment = paymentRepository.save(payment);
        log.info("Payment cancelled id={}", id);
        return payment;
    }

    @Transactional(readOnly = true)
    public Payment getPayment(UUID id) {
        return findOrThrow(id);
    }

    @Transactional(readOnly = true)
    public List<Payment> listByOrder(String orderId) {
        return paymentRepository.findByOrderId(orderId);
    }

    @Transactional(readOnly = true)
    public Page<Payment> listByCustomer(String customerId, Pageable pageable) {
        return paymentRepository.findByCustomerId(customerId, pageable);
    }

    // ── helpers ──────────────────────────────────────────────────────────────

    private Payment findOrThrow(UUID id) {
        return paymentRepository.findById(id)
                .orElseThrow(() -> new NoSuchElementException("Payment not found: " + id));
    }
}
