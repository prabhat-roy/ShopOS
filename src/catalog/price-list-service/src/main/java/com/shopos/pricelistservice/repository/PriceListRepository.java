package com.shopos.pricelistservice.repository;

import com.shopos.pricelistservice.domain.PriceList;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface PriceListRepository extends JpaRepository<PriceList, UUID> {

    Optional<PriceList> findByCode(String code);

    List<PriceList> findByActiveTrue();
}
