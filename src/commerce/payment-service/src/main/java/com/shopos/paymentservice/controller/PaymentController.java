package com.shopos.paymentservice.controller;

import com.shopos.paymentservice.domain.Payment;
import com.shopos.paymentservice.dto.CreatePaymentRequest;
import com.shopos.paymentservice.dto.RefundRequest;
import com.shopos.paymentservice.service.PaymentService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.web.PageableDefault;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping
@RequiredArgsConstructor
public class PaymentController {

    private final PaymentService paymentService;

    // ── Health ────────────────────────────────────────────────────────────────

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    // ── Payments ─────────────────────────────────────────────────────────────

    @PostMapping("/payments")
    public ResponseEntity<Payment> createPayment(
            @Valid @RequestBody CreatePaymentRequest request) {
        Payment created = paymentService.createPayment(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @GetMapping("/payments/{id}")
    public ResponseEntity<Payment> getPayment(@PathVariable UUID id) {
        return ResponseEntity.ok(paymentService.getPayment(id));
    }

    @GetMapping("/payments/order/{orderId}")
    public ResponseEntity<List<Payment>> listByOrder(@PathVariable String orderId) {
        return ResponseEntity.ok(paymentService.listByOrder(orderId));
    }

    @GetMapping("/payments/customer/{customerId}")
    public ResponseEntity<Page<Payment>> listByCustomer(
            @PathVariable String customerId,
            @PageableDefault(size = 20, sort = "createdAt") Pageable pageable) {
        return ResponseEntity.ok(paymentService.listByCustomer(customerId, pageable));
    }

    @PostMapping("/payments/{id}/capture")
    public ResponseEntity<Payment> capturePayment(@PathVariable UUID id) {
        return ResponseEntity.ok(paymentService.capturePayment(id));
    }

    @PostMapping("/payments/{id}/refund")
    public ResponseEntity<Payment> refundPayment(
            @PathVariable UUID id,
            @Valid @RequestBody RefundRequest request) {
        return ResponseEntity.ok(paymentService.refundPayment(id, request));
    }

    @PostMapping("/payments/{id}/cancel")
    public ResponseEntity<Void> cancelPayment(@PathVariable UUID id) {
        paymentService.cancelPayment(id);
        return ResponseEntity.noContent().build();
    }
}
