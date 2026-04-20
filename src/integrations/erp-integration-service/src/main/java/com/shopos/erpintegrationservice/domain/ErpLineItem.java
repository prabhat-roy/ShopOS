package com.shopos.erpintegrationservice.domain;

import java.math.BigDecimal;

/**
 * A single line item within an ERP order, carrying both ERP-native and ShopOS identifiers.
 *
 * @param erpProductId   The product identifier as known to the ERP system.
 * @param shopOsProductId The product identifier as known to ShopOS.
 * @param sku            Stock-keeping unit code.
 * @param quantity       Number of units ordered.
 * @param unitPrice      Price per unit in the order currency.
 * @param uom            Unit of measure (e.g. "EA", "KG", "EACH").
 */
public record ErpLineItem(
        String erpProductId,
        String shopOsProductId,
        String sku,
        int quantity,
        BigDecimal unitPrice,
        String uom
) {}
