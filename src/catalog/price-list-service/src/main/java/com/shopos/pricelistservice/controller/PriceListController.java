package com.shopos.pricelistservice.controller;

import com.shopos.pricelistservice.domain.PriceList;
import com.shopos.pricelistservice.domain.PriceListEntry;
import com.shopos.pricelistservice.dto.CreatePriceListRequest;
import com.shopos.pricelistservice.dto.SetEntryRequest;
import com.shopos.pricelistservice.dto.UpdatePriceListRequest;
import com.shopos.pricelistservice.service.PriceListService;
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
public class PriceListController {

    private final PriceListService priceListService;

    // POST /price-lists — create a new price list
    @PostMapping("/price-lists")
    public ResponseEntity<PriceList> createList(@Valid @RequestBody CreatePriceListRequest req) {
        return ResponseEntity.status(HttpStatus.CREATED).body(priceListService.createList(req));
    }

    // GET /price-lists — list all active price lists
    @GetMapping("/price-lists")
    public ResponseEntity<List<PriceList>> listLists() {
        return ResponseEntity.ok(priceListService.listLists());
    }

    // GET /price-lists/{id} — get a price list by ID
    @GetMapping("/price-lists/{id}")
    public ResponseEntity<PriceList> getList(@PathVariable UUID id) {
        return ResponseEntity.ok(priceListService.getList(id));
    }

    // PATCH /price-lists/{id} — partial update (name, description, active)
    @PatchMapping("/price-lists/{id}")
    public ResponseEntity<PriceList> updateList(
            @PathVariable UUID id,
            @RequestBody UpdatePriceListRequest req) {
        return ResponseEntity.ok(priceListService.updateList(id, req));
    }

    // DELETE /price-lists/{id} — soft delete (deactivate)
    @DeleteMapping("/price-lists/{id}")
    public ResponseEntity<Void> deleteList(@PathVariable UUID id) {
        priceListService.deleteList(id);
        return ResponseEntity.noContent().build();
    }

    // PUT /price-lists/{id}/entries/{productId} — upsert price for a product in this list
    @PutMapping("/price-lists/{id}/entries/{productId}")
    public ResponseEntity<PriceListEntry> setEntry(
            @PathVariable UUID id,
            @PathVariable String productId,
            @Valid @RequestBody SetEntryRequest req) {
        // Ensure the productId in path and body are consistent
        SetEntryRequest resolvedReq = new SetEntryRequest(productId, req.price());
        return ResponseEntity.ok(priceListService.setEntry(id, resolvedReq));
    }

    // GET /price-lists/{id}/entries/{productId} — get a specific product's price in this list
    @GetMapping("/price-lists/{id}/entries/{productId}")
    public ResponseEntity<PriceListEntry> getEntry(
            @PathVariable UUID id,
            @PathVariable String productId) {
        return ResponseEntity.ok(priceListService.getEntry(id, productId));
    }

    // GET /price-lists/{id}/entries — list all entries in a price list
    @GetMapping("/price-lists/{id}/entries")
    public ResponseEntity<List<PriceListEntry>> listEntries(@PathVariable UUID id) {
        return ResponseEntity.ok(priceListService.listEntries(id));
    }

    // GET /price-lists/lookup?code=WHOLESALE&productId=prod-001 — B2B price lookup by list code
    @GetMapping("/price-lists/lookup")
    public ResponseEntity<PriceListEntry> lookup(
            @RequestParam String code,
            @RequestParam String productId) {
        return ResponseEntity.ok(priceListService.getProductPrice(code, productId));
    }

    // GET /healthz — liveness probe
    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
