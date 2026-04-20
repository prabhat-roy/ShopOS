package com.shopos.pricingservice.controller;

import com.shopos.pricingservice.domain.Price;
import com.shopos.pricingservice.domain.PriceCalculation;
import com.shopos.pricingservice.dto.CalculateRequest;
import com.shopos.pricingservice.dto.SetPriceRequest;
import com.shopos.pricingservice.service.PricingService;
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
public class PricingController {

    private final PricingService pricingService;

    // POST /prices — create or update a price record
    @PostMapping("/prices")
    public ResponseEntity<Price> setPrice(@Valid @RequestBody SetPriceRequest req) {
        Price saved = pricingService.setPrice(req);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }

    // GET /prices/{productId} — get all active prices for a product (default currency USD)
    @GetMapping("/prices/{productId}")
    public ResponseEntity<List<Price>> getPrice(
            @PathVariable String productId,
            @RequestParam(defaultValue = "USD") String currency) {
        return ResponseEntity.ok(pricingService.getPrice(productId, currency));
    }

    // GET /prices/{productId}/calculate?quantity=&segment= — calculate effective price
    @GetMapping("/prices/{productId}/calculate")
    public ResponseEntity<PriceCalculation> calculate(
            @PathVariable String productId,
            @RequestParam(defaultValue = "1") int quantity,
            @RequestParam(defaultValue = "all") String segment) {
        CalculateRequest req = new CalculateRequest(productId, quantity, segment);
        return ResponseEntity.ok(pricingService.calculate(req));
    }

    // GET /prices/{productId}/all — list all price records (active + inactive)
    @GetMapping("/prices/{productId}/all")
    public ResponseEntity<List<Price>> listPrices(@PathVariable String productId) {
        return ResponseEntity.ok(pricingService.listPrices(productId));
    }

    // DELETE /prices/{id} — deactivate a price record
    @DeleteMapping("/prices/{id}")
    public ResponseEntity<Void> deletePrice(@PathVariable UUID id) {
        pricingService.deletePrice(id);
        return ResponseEntity.noContent().build();
    }

    // GET /healthz — liveness probe
    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
