package com.shopos.pricingservice.repository;

import com.shopos.pricingservice.domain.Price;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface PriceRepository extends JpaRepository<Price, UUID> {

    List<Price> findByProductIdAndActiveTrue(String productId);

    /**
     * Returns tiered price records for a product matching a given segment and minimum quantity threshold,
     * ordered from the highest qualifying tier down so the first result is the best applicable tier.
     */
    List<Price> findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
            String productId,
            String segment,
            int minQty
    );

    Optional<Price> findByProductIdAndCurrencyAndActiveTrueAndSegment(
            String productId,
            String currency,
            String segment
    );
}
