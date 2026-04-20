package com.shopos.marketplacesellerservice.domain;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.UUID;

/**
 * JPA entity for a product listing owned by a marketplace seller.
 */
@Entity
@Table(name = "seller_products", indexes = {
        @Index(name = "idx_sp_seller_id", columnList = "seller_id"),
        @Index(name = "idx_sp_product_id", columnList = "product_id"),
        @Index(name = "idx_sp_seller_sku", columnList = "seller_id, seller_sku", unique = true)
})
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class SellerProduct {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(nullable = false, updatable = false)
    private UUID id;

    /**
     * FK to the {@link Seller} entity.
     */
    @Column(name = "seller_id", nullable = false)
    private UUID sellerId;

    /**
     * Canonical product ID from the catalog domain.
     */
    @Column(name = "product_id", nullable = false, length = 100)
    private String productId;

    /**
     * Catalog SKU (from catalog-service).
     */
    @Column(name = "sku", nullable = false, length = 100)
    private String sku;

    /**
     * Seller's own internal SKU for this product.
     */
    @Column(name = "seller_sku", nullable = false, length = 100)
    private String sellerSku;

    @Column(name = "listing_price", nullable = false, precision = 19, scale = 2)
    private BigDecimal listingPrice;

    /**
     * Listing lifecycle state: ACTIVE, INACTIVE, PENDING.
     */
    @Column(name = "status", nullable = false, length = 20)
    @Builder.Default
    private String status = "PENDING";

    @Column(name = "stock_quantity")
    @Builder.Default
    private int stockQuantity = 0;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;
}
