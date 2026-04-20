package com.shopos.pricelistservice.repository;

import com.shopos.pricelistservice.domain.PriceListEntry;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface PriceListEntryRepository extends JpaRepository<PriceListEntry, UUID> {

    Optional<PriceListEntry> findByPriceListIdAndProductId(UUID priceListId, String productId);

    List<PriceListEntry> findByPriceListId(UUID priceListId);
}
