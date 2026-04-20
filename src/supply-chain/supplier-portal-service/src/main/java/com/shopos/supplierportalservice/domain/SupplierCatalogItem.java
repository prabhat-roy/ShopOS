package com.shopos.supplierportalservice.domain;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(
    name = "supplier_catalog_items",
    uniqueConstraints = @UniqueConstraint(
        name = "uq_vendor_sku",
        columnNames = {"vendor_id", "sku"}
    )
)
@Getter
@Setter
@NoArgsConstructor
public class SupplierCatalogItem {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "vendor_id", nullable = false)
    private UUID vendorId;

    @Column(name = "product_id", nullable = false, length = 200)
    private String productId;

    @Column(name = "sku", nullable = false, length = 100)
    private String sku;

    @Column(name = "product_name", nullable = false, length = 500)
    private String productName;

    @Column(name = "unit_price", nullable = false, precision = 19, scale = 4)
    private BigDecimal unitPrice;

    @Column(name = "currency", nullable = false, length = 3)
    private String currency = "USD";

    @Column(name = "min_order_qty", nullable = false)
    private int minOrderQty = 1;

    @Column(name = "lead_time_days", nullable = false)
    private int leadTimeDays;

    @Column(name = "active", nullable = false)
    private boolean active = true;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;

    public SupplierCatalogItem(
            UUID vendorId,
            String productId,
            String sku,
            String productName,
            BigDecimal unitPrice,
            String currency,
            int minOrderQty,
            int leadTimeDays) {
        this.vendorId = vendorId;
        this.productId = productId;
        this.sku = sku;
        this.productName = productName;
        this.unitPrice = unitPrice;
        this.currency = currency != null ? currency : "USD";
        this.minOrderQty = Math.max(1, minOrderQty);
        this.leadTimeDays = leadTimeDays;
        this.active = true;
    }
}
