package com.shopos.erpintegrationservice.dto;

import com.shopos.erpintegrationservice.domain.ErpSystem;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.util.Map;

/**
 * Request payload to synchronise a ShopOS order into an ERP system.
 *
 * @param shopOsOrderId   Unique ShopOS order identifier.
 * @param erpSystem       Target ERP system to push the order into.
 * @param customerMapping Key/value pairs mapping ShopOS customer fields to ERP customer master fields
 *                        (e.g. {"shopOsCustomerId": "C10042"}).
 */
public record SyncOrderRequest(
        @NotBlank(message = "shopOsOrderId must not be blank")
        String shopOsOrderId,

        @NotNull(message = "erpSystem must not be null")
        ErpSystem erpSystem,

        Map<String, String> customerMapping
) {}
