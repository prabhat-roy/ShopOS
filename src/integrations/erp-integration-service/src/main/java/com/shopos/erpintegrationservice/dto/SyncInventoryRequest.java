package com.shopos.erpintegrationservice.dto;

import com.shopos.erpintegrationservice.domain.ErpSystem;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.util.List;

/**
 * Request payload to synchronise inventory for one or more products in a warehouse.
 *
 * @param warehouseId ShopOS warehouse identifier.
 * @param erpSystem   Target ERP system.
 * @param productIds  List of ShopOS product IDs whose inventory should be synchronised.
 *                    An empty list triggers a full-warehouse sync.
 */
public record SyncInventoryRequest(
        @NotBlank(message = "warehouseId must not be blank")
        String warehouseId,

        @NotNull(message = "erpSystem must not be null")
        ErpSystem erpSystem,

        List<String> productIds
) {}
