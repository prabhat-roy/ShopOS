package com.shopos.ediservice.domain;

/**
 * Supported EDI interchange formats.
 */
public enum EdiFormat {
    /**
     * ANSI X12 — dominant in North America (orders, invoices, ASNs).
     */
    X12,

    /**
     * EDIFACT — UN/CEFACT standard dominant in Europe and international trade.
     */
    EDIFACT
}
