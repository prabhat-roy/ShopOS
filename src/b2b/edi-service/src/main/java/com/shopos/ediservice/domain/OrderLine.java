package com.shopos.ediservice.domain;

import java.math.BigDecimal;

/**
 * A single line item within a purchase order.
 *
 * @param lineNumber  Sequential line number (matches PO1 loop sequence number).
 * @param productId   Buyer's product identifier (buyer part number).
 * @param sku         Vendor's stock-keeping unit.
 * @param quantity    Ordered quantity.
 * @param unitPrice   Price per unit of measure.
 * @param uom         Unit of measure code (e.g. EA=each, CA=case, BX=box).
 */
public record OrderLine(
        int lineNumber,
        String productId,
        String sku,
        BigDecimal quantity,
        BigDecimal unitPrice,
        String uom
) {}
