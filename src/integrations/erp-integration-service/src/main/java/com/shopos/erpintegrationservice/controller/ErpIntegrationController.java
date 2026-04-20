package com.shopos.erpintegrationservice.controller;

import com.shopos.erpintegrationservice.domain.ErpSystem;
import com.shopos.erpintegrationservice.domain.SyncStatus;
import com.shopos.erpintegrationservice.dto.SyncInventoryRequest;
import com.shopos.erpintegrationservice.dto.SyncOrderRequest;
import com.shopos.erpintegrationservice.dto.SyncResponse;
import com.shopos.erpintegrationservice.service.ErpSyncService;
import com.shopos.erpintegrationservice.service.ErpTranslator;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;

/**
 * REST controller exposing ERP integration endpoints.
 *
 * <p>All write operations return HTTP 201 Created; read operations return 200 OK.
 * The {@code /healthz} endpoint is intentionally lightweight and does not touch
 * any external dependency, ensuring fast Kubernetes readiness checks.
 */
@RestController
@RequestMapping
@RequiredArgsConstructor
public class ErpIntegrationController {

    private final ErpSyncService erpSyncService;
    private final ErpTranslator erpTranslator;

    /**
     * Synchronise a ShopOS order into an ERP system.
     *
     * @param request Body containing shopOsOrderId, erpSystem, and optional customerMapping.
     * @return 201 Created with a {@link SyncResponse}.
     */
    @PostMapping("/erp/orders/sync")
    public ResponseEntity<SyncResponse> syncOrder(@Valid @RequestBody SyncOrderRequest request) {
        SyncResponse response = erpSyncService.syncOrder(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    /**
     * Synchronise inventory levels into an ERP system.
     *
     * @param request Body containing warehouseId, erpSystem, and optional list of productIds.
     * @return 201 Created with a {@link SyncResponse}.
     */
    @PostMapping("/erp/inventory/sync")
    public ResponseEntity<SyncResponse> syncInventory(@Valid @RequestBody SyncInventoryRequest request) {
        SyncResponse response = erpSyncService.syncInventory(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    /**
     * Poll the status of a previously initiated sync operation.
     *
     * @param syncId UUID returned in the original {@link SyncResponse}.
     * @return 200 OK with the {@link SyncStatus}, or 404 if the ID is unknown.
     */
    @GetMapping("/erp/sync/{syncId}")
    public ResponseEntity<Map<String, Object>> getSyncStatus(@PathVariable UUID syncId) {
        Optional<SyncStatus> status = erpSyncService.getSyncStatus(syncId);
        if (status.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.NOT_FOUND,
                    "Sync ID not found: " + syncId);
        }
        return ResponseEntity.ok(Map.of("syncId", syncId.toString(), "status", status.get().name()));
    }

    /**
     * List recent sync operations, optionally filtered by ERP system.
     *
     * @param erpSystem Optional ERP system filter. Defaults to {@code GENERIC} if not provided.
     * @param limit     Maximum number of records to return. Defaults to 20.
     * @return 200 OK with a list of {@link SyncResponse}s.
     */
    @GetMapping("/erp/syncs")
    public ResponseEntity<List<SyncResponse>> getRecentSyncs(
            @RequestParam(defaultValue = "GENERIC") String erpSystem,
            @RequestParam(defaultValue = "20") int limit) {

        ErpSystem system;
        try {
            system = ErpSystem.valueOf(erpSystem.toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST,
                    "Unknown ERP system: " + erpSystem + ". Valid values: SAP, ORACLE, NETSUITE, DYNAMICS, GENERIC");
        }

        return ResponseEntity.ok(erpSyncService.getRecentSyncs(system, limit));
    }

    /**
     * Returns the field mapping between ShopOS canonical names and ERP-native field names
     * for the specified ERP system.
     *
     * @param erpSystem ERP system identifier (case-insensitive).
     * @return 200 OK with a map of { shopOsField → erpField }.
     */
    @GetMapping("/erp/field-mappings/{erpSystem}")
    public ResponseEntity<Map<String, String>> getFieldMappings(@PathVariable String erpSystem) {
        ErpSystem system;
        try {
            system = ErpSystem.valueOf(erpSystem.toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST,
                    "Unknown ERP system: " + erpSystem + ". Valid values: SAP, ORACLE, NETSUITE, DYNAMICS, GENERIC");
        }
        return ResponseEntity.ok(erpTranslator.getFieldMapping(system));
    }

    /**
     * Kubernetes/load-balancer health check endpoint.
     *
     * @return 200 OK with {@code {"status":"ok"}}.
     */
    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> healthz() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
