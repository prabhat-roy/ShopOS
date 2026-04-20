package com.shopos.erpintegrationservice.domain;

/**
 * Enumeration of supported ERP systems.
 * Each value corresponds to a distinct field-mapping strategy in ErpTranslator.
 */
public enum ErpSystem {
    SAP,
    ORACLE,
    NETSUITE,
    DYNAMICS,
    GENERIC
}
