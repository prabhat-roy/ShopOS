package com.shopos.marketplacesellerservice.repository;

import com.shopos.marketplacesellerservice.domain.Seller;
import com.shopos.marketplacesellerservice.domain.SellerStatus;
import com.shopos.marketplacesellerservice.domain.SellerTier;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * JPA repository for {@link Seller} entities.
 */
@Repository
public interface SellerRepository extends JpaRepository<Seller, UUID> {

    /**
     * Finds the seller associated with a given organisation.
     */
    Optional<Seller> findByOrgId(UUID orgId);

    /**
     * Returns all sellers in a given lifecycle status.
     */
    List<Seller> findByStatus(SellerStatus status);

    /**
     * Returns all sellers at a given performance tier.
     */
    List<Seller> findByTier(SellerTier tier);

    /**
     * Returns sellers filtered by both status and tier.
     */
    List<Seller> findByStatusAndTier(SellerStatus status, SellerTier tier);
}
