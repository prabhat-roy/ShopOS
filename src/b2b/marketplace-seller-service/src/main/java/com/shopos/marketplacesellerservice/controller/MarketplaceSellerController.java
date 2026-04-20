package com.shopos.marketplacesellerservice.controller;

import com.shopos.marketplacesellerservice.domain.SellerProduct;
import com.shopos.marketplacesellerservice.domain.SellerStatus;
import com.shopos.marketplacesellerservice.domain.SellerTier;
import com.shopos.marketplacesellerservice.dto.CreateSellerRequest;
import com.shopos.marketplacesellerservice.dto.ListingRequest;
import com.shopos.marketplacesellerservice.dto.SellerResponse;
import com.shopos.marketplacesellerservice.service.MarketplaceSellerService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * REST controller for marketplace seller management.
 *
 * <pre>
 * POST   /sellers                             — onboard new seller
 * GET    /sellers                             — list sellers (filter by status, tier)
 * GET    /sellers/{id}                        — get seller by ID
 * GET    /sellers/by-org/{orgId}              — get seller by organisation ID
 * POST   /sellers/{id}/approve               — approve pending seller
 * POST   /sellers/{id}/suspend               — suspend active seller
 * POST   /sellers/{id}/terminate             — terminate seller
 * POST   /sellers/{id}/listings              — create product listing
 * GET    /sellers/{id}/listings              — list seller's products (filter by status)
 * GET    /sellers/{id}/listings/{listingId}  — get single listing
 * PUT    /sellers/{id}/listings/{listingId}  — update listing
 * DELETE /sellers/{id}/listings/{listingId}  — deactivate listing
 * GET    /healthz                             — health check
 * </pre>
 */
@Slf4j
@RestController
@RequiredArgsConstructor
public class MarketplaceSellerController {

    private final MarketplaceSellerService sellerService;

    // -------------------------------------------------------------------------
    // Seller CRUD
    // -------------------------------------------------------------------------

    @PostMapping(value = "/sellers",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerResponse> onboardSeller(
            @Valid @RequestBody CreateSellerRequest request) {
        log.info("POST /sellers — org={}", request.orgId());
        SellerResponse response = sellerService.onboardSeller(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping(value = "/sellers", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<List<SellerResponse>> listSellers(
            @RequestParam(required = false) SellerStatus status,
            @RequestParam(required = false) SellerTier tier) {
        log.debug("GET /sellers — status={}, tier={}", status, tier);
        return ResponseEntity.ok(sellerService.listSellers(status, tier));
    }

    @GetMapping(value = "/sellers/{id}", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerResponse> getSeller(@PathVariable UUID id) {
        return ResponseEntity.ok(sellerService.getSeller(id));
    }

    @GetMapping(value = "/sellers/by-org/{orgId}", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerResponse> getSellerByOrg(@PathVariable UUID orgId) {
        return ResponseEntity.ok(sellerService.getByOrg(orgId));
    }

    // -------------------------------------------------------------------------
    // Lifecycle transitions
    // -------------------------------------------------------------------------

    @PostMapping(value = "/sellers/{id}/approve", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerResponse> approveSeller(@PathVariable UUID id) {
        log.info("POST /sellers/{}/approve", id);
        return ResponseEntity.ok(sellerService.approveSeller(id));
    }

    @PostMapping(value = "/sellers/{id}/suspend", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerResponse> suspendSeller(@PathVariable UUID id) {
        log.info("POST /sellers/{}/suspend", id);
        return ResponseEntity.ok(sellerService.suspendSeller(id));
    }

    @PostMapping(value = "/sellers/{id}/terminate", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerResponse> terminateSeller(@PathVariable UUID id) {
        log.info("POST /sellers/{}/terminate", id);
        return ResponseEntity.ok(sellerService.terminateSeller(id));
    }

    // -------------------------------------------------------------------------
    // Product listings sub-resource
    // -------------------------------------------------------------------------

    @PostMapping(value = "/sellers/{id}/listings",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerProduct> createListing(
            @PathVariable UUID id,
            @Valid @RequestBody ListingRequest request) {
        log.info("POST /sellers/{}/listings — sku={}", id, request.sku());
        SellerProduct listing = sellerService.createListing(id, request);
        return ResponseEntity.status(HttpStatus.CREATED).body(listing);
    }

    @GetMapping(value = "/sellers/{id}/listings", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<List<SellerProduct>> listSellerProducts(
            @PathVariable UUID id,
            @RequestParam(required = false) String status) {
        return ResponseEntity.ok(sellerService.listSellerProducts(id, status));
    }

    @GetMapping(value = "/sellers/{id}/listings/{listingId}",
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerProduct> getListing(
            @PathVariable UUID id,
            @PathVariable UUID listingId) {
        return ResponseEntity.ok(sellerService.getListing(listingId));
    }

    @PutMapping(value = "/sellers/{id}/listings/{listingId}",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerProduct> updateListing(
            @PathVariable UUID id,
            @PathVariable UUID listingId,
            @Valid @RequestBody ListingRequest request) {
        log.info("PUT /sellers/{}/listings/{}", id, listingId);
        return ResponseEntity.ok(sellerService.updateListing(listingId, request));
    }

    @DeleteMapping(value = "/sellers/{id}/listings/{listingId}",
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<SellerProduct> deactivateListing(
            @PathVariable UUID id,
            @PathVariable UUID listingId) {
        log.info("DELETE /sellers/{}/listings/{}", id, listingId);
        return ResponseEntity.ok(sellerService.deactivateListing(listingId));
    }

    // -------------------------------------------------------------------------
    // Health check
    // -------------------------------------------------------------------------

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> healthz() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
