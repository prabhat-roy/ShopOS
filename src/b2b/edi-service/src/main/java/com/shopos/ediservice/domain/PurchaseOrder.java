package com.shopos.ediservice.domain;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.List;

/**
 * Business-level representation of an EDI 850 Purchase Order.
 *
 * @param poNumber    Buyer's purchase order number (BEG03).
 * @param buyer       Buyer party name (N1*BY loop).
 * @param vendor      Vendor / seller party name (N1*SE loop).
 * @param orderDate   Date the PO was issued (BEG05).
 * @param items       Line items (PO1 loops).
 * @param currency    ISO-4217 currency code, default USD.
 * @param totalAmount Computed total monetary amount.
 */
public record PurchaseOrder(
        String poNumber,
        String buyer,
        String vendor,
        LocalDate orderDate,
        List<OrderLine> items,
        String currency,
        BigDecimal totalAmount
) {

    /**
     * Convenience factory for building a PO with auto-calculated total.
     */
    public static PurchaseOrder of(
            String poNumber,
            String buyer,
            String vendor,
            LocalDate orderDate,
            List<OrderLine> items,
            String currency) {

        BigDecimal total = items.stream()
                .map(l -> l.unitPrice().multiply(l.quantity()))
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        return new PurchaseOrder(poNumber, buyer, vendor, orderDate, items, currency, total);
    }
}
