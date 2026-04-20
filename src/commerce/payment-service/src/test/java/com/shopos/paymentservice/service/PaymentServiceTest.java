package com.shopos.paymentservice.service;

import com.shopos.paymentservice.domain.Payment;
import com.shopos.paymentservice.domain.PaymentStatus;
import com.shopos.paymentservice.dto.CreatePaymentRequest;
import com.shopos.paymentservice.dto.RefundRequest;
import com.shopos.paymentservice.event.PaymentEventPublisher;
import com.shopos.paymentservice.repository.PaymentRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.NoSuchElementException;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class PaymentServiceTest {

    @Mock
    private PaymentRepository paymentRepository;

    @Mock
    private PaymentEventPublisher eventPublisher;

    @InjectMocks
    private PaymentService paymentService;

    private UUID paymentId;
    private Payment basePayment;

    @BeforeEach
    void setUp() {
        paymentId = UUID.randomUUID();
        basePayment = Payment.builder()
                .id(paymentId)
                .orderId("order-001")
                .customerId("cust-001")
                .amount(new BigDecimal("150.00"))
                .currency("USD")
                .provider("stripe")
                .status(PaymentStatus.PENDING)
                .build();
    }

    // ── createPayment ─────────────────────────────────────────────────────────

    @Test
    @DisplayName("createPayment: saves PENDING then sets AUTHORIZED and publishes processed event")
    void createPayment_setsAuthorizedAndPublishesProcessed() {
        CreatePaymentRequest req = new CreatePaymentRequest(
                "order-001", "cust-001", new BigDecimal("150.00"),
                "USD", "stripe", "tok_test_123"
        );

        // First save returns PENDING payment; second save returns AUTHORIZED payment
        Payment pendingPayment = Payment.builder()
                .id(paymentId).orderId(req.orderId()).customerId(req.customerId())
                .amount(req.amount()).currency(req.currency()).provider(req.provider())
                .status(PaymentStatus.PENDING).build();

        Payment authorizedPayment = Payment.builder()
                .id(paymentId).orderId(req.orderId()).customerId(req.customerId())
                .amount(req.amount()).currency(req.currency()).provider(req.provider())
                .status(PaymentStatus.AUTHORIZED).providerTxId("sim_abc123").build();

        when(paymentRepository.save(any(Payment.class)))
                .thenReturn(pendingPayment)
                .thenReturn(authorizedPayment);

        Payment result = paymentService.createPayment(req);

        assertThat(result.getStatus()).isEqualTo(PaymentStatus.AUTHORIZED);
        verify(paymentRepository, times(2)).save(any(Payment.class));
        verify(eventPublisher).publishProcessed(authorizedPayment);
    }

    // ── capturePayment ────────────────────────────────────────────────────────

    @Test
    @DisplayName("capturePayment: transitions AUTHORIZED → CAPTURED")
    void capturePayment_fromAuthorized_succeeds() {
        Payment authorizedPayment = Payment.builder()
                .id(paymentId).orderId("order-001").customerId("cust-001")
                .amount(new BigDecimal("150.00")).currency("USD")
                .status(PaymentStatus.AUTHORIZED).build();

        Payment capturedPayment = Payment.builder()
                .id(paymentId).orderId("order-001").customerId("cust-001")
                .amount(new BigDecimal("150.00")).currency("USD")
                .status(PaymentStatus.CAPTURED).build();

        when(paymentRepository.findById(paymentId)).thenReturn(Optional.of(authorizedPayment));
        when(paymentRepository.save(any(Payment.class))).thenReturn(capturedPayment);

        Payment result = paymentService.capturePayment(paymentId);

        assertThat(result.getStatus()).isEqualTo(PaymentStatus.CAPTURED);
        verify(paymentRepository).save(any(Payment.class));
    }

    @Test
    @DisplayName("capturePayment: throws IllegalStateException when status is not AUTHORIZED")
    void capturePayment_fromPending_throwsIllegalState() {
        when(paymentRepository.findById(paymentId)).thenReturn(Optional.of(basePayment));

        assertThatThrownBy(() -> paymentService.capturePayment(paymentId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("cannot be captured");

        verify(paymentRepository, never()).save(any());
    }

    @Test
    @DisplayName("capturePayment: throws NoSuchElementException for unknown id")
    void capturePayment_unknownId_throwsNotFound() {
        when(paymentRepository.findById(paymentId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> paymentService.capturePayment(paymentId))
                .isInstanceOf(NoSuchElementException.class)
                .hasMessageContaining("Payment not found");
    }

    // ── refundPayment ─────────────────────────────────────────────────────────

    @Test
    @DisplayName("refundPayment: transitions CAPTURED → REFUNDED for valid amount")
    void refundPayment_validAmount_succeeds() {
        Payment capturedPayment = Payment.builder()
                .id(paymentId).orderId("order-001").customerId("cust-001")
                .amount(new BigDecimal("150.00")).currency("USD")
                .status(PaymentStatus.CAPTURED).build();

        Payment refundedPayment = Payment.builder()
                .id(paymentId).orderId("order-001").customerId("cust-001")
                .amount(new BigDecimal("150.00")).currency("USD")
                .status(PaymentStatus.REFUNDED).build();

        when(paymentRepository.findById(paymentId)).thenReturn(Optional.of(capturedPayment));
        when(paymentRepository.save(any(Payment.class))).thenReturn(refundedPayment);

        RefundRequest req = new RefundRequest(new BigDecimal("50.00"), "Customer request");
        Payment result = paymentService.refundPayment(paymentId, req);

        assertThat(result.getStatus()).isEqualTo(PaymentStatus.REFUNDED);
        verify(eventPublisher).publishProcessed(refundedPayment);
        verify(eventPublisher, never()).publishFailed(any(), any());
    }

    @Test
    @DisplayName("refundPayment: sets FAILED and publishes failed event when amount exceeds original")
    void refundPayment_exceedsOriginalAmount_failsAndPublishesFailed() {
        Payment capturedPayment = Payment.builder()
                .id(paymentId).orderId("order-001").customerId("cust-001")
                .amount(new BigDecimal("150.00")).currency("USD")
                .status(PaymentStatus.CAPTURED).build();

        Payment failedPayment = Payment.builder()
                .id(paymentId).orderId("order-001").customerId("cust-001")
                .amount(new BigDecimal("150.00")).currency("USD")
                .status(PaymentStatus.FAILED).build();

        when(paymentRepository.findById(paymentId)).thenReturn(Optional.of(capturedPayment));
        when(paymentRepository.save(any(Payment.class))).thenReturn(failedPayment);

        RefundRequest req = new RefundRequest(new BigDecimal("999.00"), "Too much");
        Payment result = paymentService.refundPayment(paymentId, req);

        assertThat(result.getStatus()).isEqualTo(PaymentStatus.FAILED);
        verify(eventPublisher).publishFailed(eq(failedPayment), anyString());
        verify(eventPublisher, never()).publishProcessed(any());
    }

    @Test
    @DisplayName("refundPayment: throws IllegalStateException when status is not CAPTURED")
    void refundPayment_notCaptured_throwsIllegalState() {
        when(paymentRepository.findById(paymentId)).thenReturn(Optional.of(basePayment));

        RefundRequest req = new RefundRequest(new BigDecimal("50.00"), "reason");

        assertThatThrownBy(() -> paymentService.refundPayment(paymentId, req))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("cannot be refunded");
    }

    // ── cancelPayment ─────────────────────────────────────────────────────────

    @Test
    @DisplayName("cancelPayment: transitions PENDING → CANCELLED")
    void cancelPayment_fromPending_succeeds() {
        Payment cancelledPayment = Payment.builder()
                .id(paymentId).orderId("order-001").customerId("cust-001")
                .amount(new BigDecimal("150.00")).currency("USD")
                .status(PaymentStatus.CANCELLED).build();

        when(paymentRepository.findById(paymentId)).thenReturn(Optional.of(basePayment));
        when(paymentRepository.save(any(Payment.class))).thenReturn(cancelledPayment);

        Payment result = paymentService.cancelPayment(paymentId);

        assertThat(result.getStatus()).isEqualTo(PaymentStatus.CANCELLED);
    }

    @Test
    @DisplayName("cancelPayment: throws IllegalStateException from CAPTURED status")
    void cancelPayment_fromCaptured_throwsIllegalState() {
        Payment capturedPayment = Payment.builder()
                .id(paymentId).status(PaymentStatus.CAPTURED)
                .amount(new BigDecimal("150.00")).build();

        when(paymentRepository.findById(paymentId)).thenReturn(Optional.of(capturedPayment));

        assertThatThrownBy(() -> paymentService.cancelPayment(paymentId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("cannot be cancelled");
    }
}
