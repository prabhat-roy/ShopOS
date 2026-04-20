package com.shopos.supplierportalservice.controller;

import com.shopos.supplierportalservice.domain.SupplierInvoiceStatus;
import com.shopos.supplierportalservice.dto.*;
import com.shopos.supplierportalservice.service.SupplierPortalService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

@Slf4j
@RestController
@RequiredArgsConstructor
public class SupplierPortalController {

    private final SupplierPortalService service;

    // ------------------------------------------------------------------
    // Health
    // ------------------------------------------------------------------

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> healthz() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    // ------------------------------------------------------------------
    // Invoice endpoints
    // ------------------------------------------------------------------

    @PostMapping("/invoices")
    public ResponseEntity<InvoiceResponse> createInvoice(
            @Valid @RequestBody CreateInvoiceRequest req) {
        return ResponseEntity
                .status(HttpStatus.CREATED)
                .body(InvoiceResponse.from(service.createInvoice(req)));
    }

    @GetMapping("/invoices")
    public ResponseEntity<List<InvoiceResponse>> listInvoices(
            @RequestParam(required = false) UUID vendorId,
            @RequestParam(required = false) SupplierInvoiceStatus status) {
        List<InvoiceResponse> results = service.listInvoices(vendorId, status)
                .stream()
                .map(InvoiceResponse::from)
                .collect(Collectors.toList());
        return ResponseEntity.ok(results);
    }

    @GetMapping("/invoices/{id}")
    public ResponseEntity<InvoiceResponse> getInvoice(@PathVariable UUID id) {
        return ResponseEntity.ok(InvoiceResponse.from(service.getInvoice(id)));
    }

    @PostMapping("/invoices/{id}/submit")
    public ResponseEntity<InvoiceResponse> submitInvoice(@PathVariable UUID id) {
        return ResponseEntity.ok(InvoiceResponse.from(service.submitInvoice(id)));
    }

    @PostMapping("/invoices/{id}/approve")
    public ResponseEntity<InvoiceResponse> approveInvoice(@PathVariable UUID id) {
        return ResponseEntity.ok(InvoiceResponse.from(service.approveInvoice(id)));
    }

    @PostMapping("/invoices/{id}/reject")
    public ResponseEntity<InvoiceResponse> rejectInvoice(
            @PathVariable UUID id,
            @RequestBody Map<String, String> body) {
        String reason = body.getOrDefault("reason", "No reason provided");
        return ResponseEntity.ok(InvoiceResponse.from(service.rejectInvoice(id, reason)));
    }

    @PostMapping("/invoices/{id}/pay")
    public ResponseEntity<InvoiceResponse> markPaid(@PathVariable UUID id) {
        return ResponseEntity.ok(InvoiceResponse.from(service.markPaid(id)));
    }

    // ------------------------------------------------------------------
    // Catalog endpoints
    // ------------------------------------------------------------------

    @PutMapping("/catalog")
    public ResponseEntity<CatalogItemResponse> upsertCatalogItem(
            @Valid @RequestBody UpsertCatalogItemRequest req) {
        CatalogItemResponse resp = CatalogItemResponse.from(service.upsertCatalogItem(req));
        return ResponseEntity.ok(resp);
    }

    @GetMapping("/catalog")
    public ResponseEntity<List<CatalogItemResponse>> listCatalogItems(
            @RequestParam(required = false) UUID vendorId,
            @RequestParam(defaultValue = "false") boolean activeOnly) {
        List<CatalogItemResponse> results = service.listCatalogItems(vendorId, activeOnly)
                .stream()
                .map(CatalogItemResponse::from)
                .collect(Collectors.toList());
        return ResponseEntity.ok(results);
    }

    @GetMapping("/catalog/{id}")
    public ResponseEntity<CatalogItemResponse> getCatalogItem(@PathVariable UUID id) {
        return ResponseEntity.ok(CatalogItemResponse.from(service.getCatalogItem(id)));
    }

    @DeleteMapping("/catalog/{id}")
    public ResponseEntity<CatalogItemResponse> deactivateCatalogItem(@PathVariable UUID id) {
        return ResponseEntity.ok(CatalogItemResponse.from(service.deactivateCatalogItem(id)));
    }
}
