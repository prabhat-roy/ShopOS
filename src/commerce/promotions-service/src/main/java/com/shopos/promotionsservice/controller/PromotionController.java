package com.shopos.promotionsservice.controller;

import com.shopos.promotionsservice.domain.Promotion;
import com.shopos.promotionsservice.dto.ApplyCouponRequest;
import com.shopos.promotionsservice.dto.CreatePromotionRequest;
import com.shopos.promotionsservice.dto.ValidateRequest;
import com.shopos.promotionsservice.dto.ValidateResponse;
import com.shopos.promotionsservice.service.PromotionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping
@RequiredArgsConstructor
public class PromotionController {

    private final PromotionService promotionService;

    // ── Health ────────────────────────────────────────────────────────────────

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    // ── Promotions ────────────────────────────────────────────────────────────

    @PostMapping("/promotions")
    public ResponseEntity<Promotion> createPromotion(
            @Valid @RequestBody CreatePromotionRequest request) {
        Promotion created = promotionService.createPromotion(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @GetMapping("/promotions")
    public ResponseEntity<List<Promotion>> listPromotions(
            @RequestParam(value = "activeOnly", defaultValue = "false") boolean activeOnly) {
        return ResponseEntity.ok(promotionService.listPromotions(activeOnly));
    }

    @GetMapping("/promotions/{id}")
    public ResponseEntity<Promotion> getPromotion(@PathVariable UUID id) {
        return ResponseEntity.ok(promotionService.getPromotion(id));
    }

    @DeleteMapping("/promotions/{id}")
    public ResponseEntity<Void> deactivatePromotion(@PathVariable UUID id) {
        promotionService.deactivate(id);
        return ResponseEntity.noContent().build();
    }

    // ── Validate / Apply ──────────────────────────────────────────────────────

    @PostMapping("/promotions/validate")
    public ResponseEntity<ValidateResponse> validateCoupon(
            @Valid @RequestBody ValidateRequest request) {
        ValidateResponse response = promotionService.validateCoupon(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/promotions/apply")
    public ResponseEntity<Map<String, Object>> applyCoupon(
            @Valid @RequestBody ApplyCouponRequest request) {
        BigDecimal discount = promotionService.applyCoupon(request.code());
        return ResponseEntity.ok(Map.of(
                "code", request.code().toUpperCase().strip(),
                "discountAmount", discount
        ));
    }
}
