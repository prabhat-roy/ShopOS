package com.shopos.marketplacesellerservice.repository;

import com.shopos.marketplacesellerservice.domain.SellerProduct;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * JPA repository for {@link SellerProduct} (product listing) entities.
 */
@Repository
public interface SellerProductRepository extends JpaRepository<SellerProduct, UUID> {

    /**
     * Returns all listings for a seller, regardless of status.
     */
    List<SellerProduct> findBySellerId(UUID sellerId);

    /**
     * Returns listings for a seller filtered by lifecycle status.
     *
     * @param status ACTIVE | INACTIVE | PENDING
     */
    List<SellerProduct> findBySellerIdAndStatus(UUID sellerId, String status);

    /**
     * Returns all seller listings for a canonical catalog product ID.
     */
    List<SellerProduct> findByProductId(String productId);

    /**
     * Finds a seller's listing by its catalog SKU.
     */
    Optional<SellerProduct> findBySellerIdAndSku(UUID sellerId, String sku);

    /**
     * Finds a listing by seller and seller-specific SKU.
     */
    Optional<SellerProduct> findBySellerIdAndSellerSku(UUID sellerId, String sellerSku);
}
