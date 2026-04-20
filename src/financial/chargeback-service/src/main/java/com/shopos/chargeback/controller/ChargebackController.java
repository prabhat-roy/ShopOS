package com.shopos.chargeback.controller;

import com.shopos.chargeback.domain.ChargebackStatus;
import com.shopos.chargeback.dto.ChargebackResponse;
import com.shopos.chargeback.dto.CreateChargebackRequest;
import com.shopos.chargeback.service.ChargebackService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/chargebacks")
@RequiredArgsConstructor
public class ChargebackController {

    private final ChargebackService chargebackService;

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    @PostMapping
    public ResponseEntity<ChargebackResponse> create(@Valid @RequestBody CreateChargebackRequest request) {
        return ResponseEntity.status(HttpStatus.CREATED).body(chargebackService.createChargeback(request));
    }

    @GetMapping("/{id}")
    public ResponseEntity<ChargebackResponse> getById(@PathVariable UUID id) {
        return ResponseEntity.ok(chargebackService.getChargeback(id));
    }

    @GetMapping("/customer/{customerId}")
    public ResponseEntity<List<ChargebackResponse>> getByCustomer(@PathVariable String customerId) {
        return ResponseEntity.ok(chargebackService.getChargebacksByCustomer(customerId));
    }

    @PostMapping("/{id}/evidence")
    public ResponseEntity<ChargebackResponse> submitEvidence(@PathVariable UUID id) {
        return ResponseEntity.ok(chargebackService.submitEvidence(id));
    }

    @PostMapping("/{id}/resolve")
    public ResponseEntity<ChargebackResponse> resolve(
            @PathVariable UUID id,
            @RequestParam ChargebackStatus resolution) {
        return ResponseEntity.ok(chargebackService.resolveChargeback(id, resolution));
    }
}
