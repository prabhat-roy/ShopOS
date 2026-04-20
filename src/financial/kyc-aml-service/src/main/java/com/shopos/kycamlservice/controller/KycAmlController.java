package com.shopos.kycamlservice.controller;

import com.shopos.kycamlservice.dto.*;
import com.shopos.kycamlservice.service.KycAmlService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequiredArgsConstructor
public class KycAmlController {

    private final KycAmlService kycAmlService;

    // -------------------------------------------------------------------------
    // Health
    // -------------------------------------------------------------------------

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    // -------------------------------------------------------------------------
    // KYC endpoints
    // -------------------------------------------------------------------------

    /**
     * Create a new KYC record. Returns 201 Created with the created resource.
     */
    @PostMapping("/kyc")
    public ResponseEntity<KycResponse> createKycRecord(@Valid @RequestBody CreateKycRequest request) {
        KycResponse response = kycAmlService.createKycRecord(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    /**
     * Retrieve a KYC record by its own UUID.
     */
    @GetMapping("/kyc/{id}")
    public ResponseEntity<KycResponse> getKycRecord(@PathVariable UUID id) {
        return ResponseEntity.ok(kycAmlService.getKycRecord(id));
    }

    /**
     * Retrieve a KYC record by the customer UUID.
     */
    @GetMapping("/kyc/customer/{customerId}")
    public ResponseEntity<KycResponse> getByCustomerId(@PathVariable UUID customerId) {
        return ResponseEntity.ok(kycAmlService.getByCustomerId(customerId));
    }

    /**
     * Move a KYC record from PENDING → IN_PROGRESS.
     */
    @PostMapping("/kyc/{id}/start")
    public ResponseEntity<Void> startVerification(@PathVariable UUID id) {
        kycAmlService.startVerification(id);
        return ResponseEntity.noContent().build();
    }

    /**
     * Move a KYC record from IN_PROGRESS → VERIFIED.
     */
    @PostMapping("/kyc/{id}/verify")
    public ResponseEntity<Void> verifyKyc(@PathVariable UUID id) {
        kycAmlService.verifyKyc(id);
        return ResponseEntity.noContent().build();
    }

    /**
     * Move a KYC record from IN_PROGRESS → REJECTED with a mandatory reason.
     */
    @PostMapping("/kyc/{id}/reject")
    public ResponseEntity<Void> rejectKyc(
            @PathVariable UUID id,
            @Valid @RequestBody RejectKycRequest request) {
        kycAmlService.rejectKyc(id, request.reason());
        return ResponseEntity.noContent().build();
    }

    /**
     * Move a KYC record from VERIFIED → SUSPENDED.
     */
    @PostMapping("/kyc/{id}/suspend")
    public ResponseEntity<Void> suspendKyc(@PathVariable UUID id) {
        kycAmlService.suspendKyc(id);
        return ResponseEntity.noContent().build();
    }

    /**
     * Trigger the batch expiry detection immediately (administrative endpoint).
     * Returns the count of records newly marked EXPIRED.
     */
    @PostMapping("/kyc/detect-expired")
    public ResponseEntity<Map<String, Integer>> detectExpired() {
        int count = kycAmlService.detectExpiredKyc();
        return ResponseEntity.ok(Map.of("expiredCount", count));
    }

    // -------------------------------------------------------------------------
    // AML endpoints
    // -------------------------------------------------------------------------

    /**
     * Run an AML check for a customer. Returns 201 Created with the check result.
     */
    @PostMapping("/aml/checks")
    public ResponseEntity<AmlCheckResponse> runAmlCheck(@Valid @RequestBody RunAmlCheckRequest request) {
        AmlCheckResponse response = kycAmlService.runAmlCheck(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    /**
     * Retrieve all AML checks for a specific customer.
     */
    @GetMapping("/aml/checks/{customerId}")
    public ResponseEntity<List<AmlCheckResponse>> getAmlChecks(@PathVariable UUID customerId) {
        return ResponseEntity.ok(kycAmlService.getAmlChecks(customerId));
    }

    /**
     * Resolve an existing AML check (record the outcome and analyst who resolved it).
     */
    @PostMapping("/aml/checks/{id}/resolve")
    public ResponseEntity<Void> resolveAmlCheck(
            @PathVariable UUID id,
            @Valid @RequestBody ResolveAmlCheckRequest request) {
        kycAmlService.resolveAmlCheck(id, request.resolution(), request.resolvedBy());
        return ResponseEntity.noContent().build();
    }
}
