package com.shopos.erpintegrationservice.domain;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;

/**
 * Represents an order that has been translated into (or received from) the ERP system.
 *
 * @param erpOrderId      The document/order number as assigned by the ERP system.
 * @param shopOsOrderId   The original order identifier in ShopOS.
 * @param erpSystem       Which ERP system owns this record.
 * @param customerErpId   The customer master record ID in the ERP system.
 * @param lineItems       Individual product lines comprising the order.
 * @param totalAmount     Net order value.
 * @param currency        ISO 4217 currency code (e.g. "USD").
 * @param status          Current synchronisation status.
 * @param syncedAt        Timestamp when the record was last synchronised.
 */
public record ErpOrder(
        String erpOrderId,
        String shopOsOrderId,
        ErpSystem erpSystem,
        String customerErpId,
        List<ErpLineItem> lineItems,
        BigDecimal totalAmount,
        String currency,
        SyncStatus status,
        Instant syncedAt
) {}
