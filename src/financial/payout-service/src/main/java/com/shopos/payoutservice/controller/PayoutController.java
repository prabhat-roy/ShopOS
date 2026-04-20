package com.shopos.payoutservice.controller;

import com.shopos.payoutservice.domain.PayoutStatus;
import com.shopos.payoutservice.dto.CreatePayoutRequest;
import com.shopos.payoutservice.dto.PayoutResponse;
import com.shopos.payoutservice.service.PayoutService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.web.PageableDefault;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/payouts")
@RequiredArgsConstructor
public class PayoutController {

    private final PayoutService payoutService;

    /**
     * POST /payouts — creates a new payout in PENDING status.
     * Returns 201 Created with the created resource.
     */
    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public PayoutResponse createPayout(@Valid @RequestBody CreatePayoutRequest request) {
        return payoutService.createPayout(request);
    }

    /**
     * GET /payouts/{id} — retrieves a single payout by UUID.
     */
    @GetMapping("/{id}")
    public PayoutResponse getPayout(@PathVariable UUID id) {
        return payoutService.getPayout(id);
    }

    /**
     * GET /payouts?vendorId=&status=&page=&size=&sort=
     * Returns a paginated list filtered by vendorId (required) and optional status.
     * Falls back to listing all payouts with a given status when vendorId is absent.
     */
    @GetMapping
    public ResponseEntity<?> listPayouts(
            @RequestParam(required = false) UUID vendorId,
            @RequestParam(required = false) PayoutStatus status,
            @PageableDefault(size = 20, sort = "createdAt") Pageable pageable) {

        if (vendorId != null) {
            Page<PayoutResponse> page = payoutService.listByVendor(vendorId, status, pageable);
            return ResponseEntity.ok(page);
        }
        if (status != null) {
            List<PayoutResponse> list = payoutService.listByStatus(status);
            return ResponseEntity.ok(list);
        }
        return ResponseEntity.badRequest()
                .body(Map.of("error", "Provide at least one of: vendorId, status"));
    }

    /**
     * POST /payouts/{id}/process — processes a PENDING payout (→ PROCESSING → COMPLETED).
     */
    @PostMapping("/{id}/process")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void processPayout(@PathVariable UUID id) {
        payoutService.processPayout(id);
    }

    /**
     * POST /payouts/{id}/fail — marks a PROCESSING payout as FAILED.
     * Accepts optional JSON body: { "reason": "..." }
     */
    @PostMapping("/{id}/fail")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void failPayout(@PathVariable UUID id,
                           @RequestBody(required = false) Map<String, String> body) {
        String reason = (body != null) ? body.getOrDefault("reason", "Unspecified failure") : "Unspecified failure";
        payoutService.failPayout(id, reason);
    }

    /**
     * POST /payouts/{id}/cancel — cancels a PENDING payout.
     */
    @PostMapping("/{id}/cancel")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void cancelPayout(@PathVariable UUID id) {
        payoutService.cancelPayout(id);
    }

    /**
     * POST /payouts/{id}/retry — resets a FAILED payout back to PENDING.
     */
    @PostMapping("/{id}/retry")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void retryPayout(@PathVariable UUID id) {
        payoutService.retryPayout(id);
    }

    /**
     * POST /payouts/process-due — batch job endpoint.
     * Processes all PENDING payouts that are past their scheduledAt (or have no schedule).
     * Returns the count of payouts successfully processed.
     */
    @PostMapping("/process-due")
    public ResponseEntity<Map<String, Integer>> processDue() {
        int count = payoutService.processDuePayouts();
        return ResponseEntity.ok(Map.of("processedCount", count));
    }
}
