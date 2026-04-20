package com.shopos.erpintegrationservice.service;

import com.shopos.erpintegrationservice.domain.ErpInventory;
import com.shopos.erpintegrationservice.domain.ErpOrder;
import com.shopos.erpintegrationservice.domain.ErpSystem;
import com.shopos.erpintegrationservice.domain.SyncStatus;
import com.shopos.erpintegrationservice.dto.SyncInventoryRequest;
import com.shopos.erpintegrationservice.dto.SyncOrderRequest;
import com.shopos.erpintegrationservice.dto.SyncResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Deque;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.stream.StreamSupport;

/**
 * Orchestrates synchronisation operations between ShopOS and external ERP systems.
 *
 * <p>This service is intentionally stateless with regard to persistent storage; all in-flight
 * and recent sync records are held in bounded in-memory structures to support status polling
 * during the lifetime of a single service instance.  A production deployment would back
 * these structures with a database or distributed cache.
 *
 * <p>The actual ERP API call is simulated: the service builds the translated payload,
 * assigns a sync ID, and records a SUCCESS result.  In a real implementation the HTTP/RFC
 * client call would happen here, and FAILED/PARTIAL status would be set on error.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class ErpSyncService {

    /** Maximum number of sync records retained per ERP system. */
    private static final int MAX_HISTORY_PER_SYSTEM = 1000;

    /** Maximum total number of sync records retained in the global lookup map. */
    private static final int MAX_GLOBAL_RECORDS = 5000;

    private final ErpTranslator erpTranslator;

    /**
     * Global map of syncId → SyncResponse for fast status lookups.
     * Bounded to MAX_GLOBAL_RECORDS via periodic eviction on write.
     */
    private final ConcurrentHashMap<UUID, SyncResponse> syncIndex = new ConcurrentHashMap<>();

    /**
     * Per-system ordered deques of sync IDs (newest first).
     * Each deque is bounded to MAX_HISTORY_PER_SYSTEM entries.
     */
    private final ConcurrentHashMap<ErpSystem, ConcurrentLinkedDeque<UUID>> systemHistory =
            new ConcurrentHashMap<>();

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Translates and synchronises a single ShopOS order into the specified ERP system.
     *
     * @param request Order sync request containing the ShopOS order ID and ERP system target.
     * @return A {@link SyncResponse} with a unique sync ID and final status.
     */
    public SyncResponse syncOrder(SyncOrderRequest request) {
        UUID syncId = UUID.randomUUID();
        log.info("Starting order sync [syncId={}, orderId={}, erpSystem={}]",
                syncId, request.shopOsOrderId(), request.erpSystem());

        List<String> errors = new ArrayList<>();
        int recordsProcessed = 0;
        SyncStatus status;

        try {
            Map<String, Object> orderData = buildOrderPayload(request);
            ErpOrder erpOrder = erpTranslator.translateOrderToErp(orderData, request.erpSystem());

            // Simulate ERP API call — replace with actual client in production
            simulateErpCall(request.erpSystem(), "order", erpOrder.erpOrderId());

            recordsProcessed = 1;
            status = SyncStatus.SUCCESS;
            log.info("Order sync succeeded [syncId={}, erpOrderId={}]", syncId, erpOrder.erpOrderId());
        } catch (Exception e) {
            log.error("Order sync failed [syncId={}]: {}", syncId, e.getMessage(), e);
            errors.add("Order sync failed: " + e.getMessage());
            status = SyncStatus.FAILED;
        }

        SyncResponse response = new SyncResponse(syncId, status, recordsProcessed, errors, Instant.now());
        record(request.erpSystem(), syncId, response);
        return response;
    }

    /**
     * Synchronises inventory levels for one or more products from the specified warehouse.
     *
     * @param request Inventory sync request.
     * @return A {@link SyncResponse} summarising processed records and any errors.
     */
    public SyncResponse syncInventory(SyncInventoryRequest request) {
        UUID syncId = UUID.randomUUID();
        log.info("Starting inventory sync [syncId={}, warehouseId={}, erpSystem={}, products={}]",
                syncId, request.warehouseId(), request.erpSystem(),
                request.productIds() == null ? "ALL" : request.productIds().size());

        List<String> errors = new ArrayList<>();
        int recordsProcessed = 0;

        List<String> productIds = request.productIds() != null && !request.productIds().isEmpty()
                ? request.productIds()
                : generateSampleProductIds(3);

        for (String productId : productIds) {
            try {
                Map<String, Object> inventoryData = buildInventoryPayload(productId, request.warehouseId());
                ErpInventory inventory = erpTranslator.translateInventory(inventoryData, request.erpSystem());

                // Simulate ERP API call — replace with actual client in production
                simulateErpCall(request.erpSystem(), "inventory", inventory.erpProductId());
                recordsProcessed++;
            } catch (Exception e) {
                log.warn("Inventory sync failed for product [syncId={}, productId={}]: {}",
                        syncId, productId, e.getMessage());
                errors.add("Product " + productId + ": " + e.getMessage());
            }
        }

        SyncStatus status = resolveStatus(recordsProcessed, productIds.size(), errors);
        SyncResponse response = new SyncResponse(syncId, status, recordsProcessed, errors, Instant.now());
        record(request.erpSystem(), syncId, response);
        return response;
    }

    /**
     * Returns the current status of a sync operation by its ID.
     *
     * @param syncId The sync operation identifier returned at creation time.
     * @return {@link Optional} containing the {@link SyncStatus} if found, or empty if unknown.
     */
    public Optional<SyncStatus> getSyncStatus(UUID syncId) {
        SyncResponse response = syncIndex.get(syncId);
        return Optional.ofNullable(response).map(SyncResponse::status);
    }

    /**
     * Returns recent sync responses for the specified ERP system.
     *
     * @param erpSystem Target ERP system filter.
     * @param limit     Maximum number of responses to return; capped at MAX_HISTORY_PER_SYSTEM.
     * @return Ordered list of {@link SyncResponse}s, newest first.
     */
    public List<SyncResponse> getRecentSyncs(ErpSystem erpSystem, int limit) {
        int effectiveLimit = Math.min(Math.max(1, limit), MAX_HISTORY_PER_SYSTEM);
        Deque<UUID> history = systemHistory.getOrDefault(erpSystem, new ConcurrentLinkedDeque<>());

        return StreamSupport.stream(
                        ((Iterable<UUID>) history).spliterator(), false)
                .limit(effectiveLimit)
                .map(syncIndex::get)
                .filter(java.util.Objects::nonNull)
                .toList();
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    private void record(ErpSystem erpSystem, UUID syncId, SyncResponse response) {
        syncIndex.put(syncId, response);

        ConcurrentLinkedDeque<UUID> deque = systemHistory.computeIfAbsent(
                erpSystem, k -> new ConcurrentLinkedDeque<>());
        deque.addFirst(syncId);

        // Trim per-system history
        while (deque.size() > MAX_HISTORY_PER_SYSTEM) {
            UUID evicted = deque.pollLast();
            if (evicted != null) {
                syncIndex.remove(evicted);
            }
        }

        // Safety net: trim global index if it grows too large
        if (syncIndex.size() > MAX_GLOBAL_RECORDS) {
            UUID oldest = syncIndex.keys().nextElement();
            syncIndex.remove(oldest);
        }
    }

    private Map<String, Object> buildOrderPayload(SyncOrderRequest request) {
        Map<String, Object> order = new HashMap<>();
        order.put("orderId", request.shopOsOrderId());
        order.put("totalAmount", "499.99");
        order.put("currency", "USD");

        if (request.customerMapping() != null) {
            order.putAll(request.customerMapping());
        }

        Map<String, Object> item1 = new HashMap<>();
        item1.put("productId", "prod-001");
        item1.put("sku", "SKU-001");
        item1.put("quantity", 2);
        item1.put("unitPrice", "149.99");
        item1.put("uom", "EA");

        Map<String, Object> item2 = new HashMap<>();
        item2.put("productId", "prod-002");
        item2.put("sku", "SKU-002");
        item2.put("quantity", 1);
        item2.put("unitPrice", "200.01");
        item2.put("uom", "EA");

        order.put("lineItems", List.of(item1, item2));
        return order;
    }

    private Map<String, Object> buildInventoryPayload(String productId, String warehouseId) {
        Map<String, Object> inv = new HashMap<>();
        inv.put("shopOsProductId", productId);
        inv.put("sku", "SKU-" + productId.toUpperCase());
        inv.put("quantity", 100);
        inv.put("warehouseId", warehouseId);
        return inv;
    }

    private List<String> generateSampleProductIds(int count) {
        List<String> ids = new ArrayList<>();
        for (int i = 1; i <= count; i++) {
            ids.add("prod-00" + i);
        }
        return ids;
    }

    private SyncStatus resolveStatus(int processed, int total, List<String> errors) {
        if (errors.isEmpty()) return SyncStatus.SUCCESS;
        if (processed == 0) return SyncStatus.FAILED;
        return SyncStatus.PARTIAL;
    }

    /**
     * Simulates an outbound ERP API call.
     * Replace this method body with a real HTTP/RFC/BAPI client in production.
     *
     * @param erpSystem  Target system.
     * @param entityType "order" or "inventory".
     * @param erpId      The ERP-native document or material ID being pushed.
     */
    private void simulateErpCall(ErpSystem erpSystem, String entityType, String erpId) {
        log.debug("Simulated ERP call [system={}, entity={}, id={}]", erpSystem, entityType, erpId);
        // In a real service: httpClient.post(erpSystem.getBaseUrl() + "/" + entityType, payload);
    }
}
