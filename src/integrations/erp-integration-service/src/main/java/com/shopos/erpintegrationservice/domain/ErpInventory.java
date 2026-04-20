package com.shopos.erpintegrationservice.domain;

import java.time.Instant;

/**
 * Inventory position for a single product in a specific warehouse, translated for ERP consumption.
 *
 * @param erpProductId    The product identifier as known to the ERP system.
 * @param shopOsProductId The product identifier as known to ShopOS.
 * @param sku             Stock-keeping unit code.
 * @param quantity        Available stock quantity.
 * @param warehouseId     ShopOS warehouse identifier.
 * @param lastUpdated     When the inventory level was last updated in ShopOS.
 */
public record ErpInventory(
        String erpProductId,
        String shopOsProductId,
        String sku,
        int quantity,
        String warehouseId,
        Instant lastUpdated
) {}
