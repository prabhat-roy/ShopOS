package com.shopos.supplierportalservice.dto;

import com.shopos.supplierportalservice.domain.SupplierCatalogItem;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
public class CatalogItemResponse {

    private UUID id;
    private UUID vendorId;
    private String productId;
    private String sku;
    private String productName;
    private BigDecimal unitPrice;
    private String currency;
    private int minOrderQty;
    private int leadTimeDays;
    private boolean active;
    private LocalDateTime updatedAt;

    public static CatalogItemResponse from(SupplierCatalogItem item) {
        return CatalogItemResponse.builder()
                .id(item.getId())
                .vendorId(item.getVendorId())
                .productId(item.getProductId())
                .sku(item.getSku())
                .productName(item.getProductName())
                .unitPrice(item.getUnitPrice())
                .currency(item.getCurrency())
                .minOrderQty(item.getMinOrderQty())
                .leadTimeDays(item.getLeadTimeDays())
                .active(item.isActive())
                .updatedAt(item.getUpdatedAt())
                .build();
    }
}
