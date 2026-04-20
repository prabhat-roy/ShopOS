package com.shopos.promotionsservice.repository;

import com.shopos.promotionsservice.domain.Promotion;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface PromotionRepository extends JpaRepository<Promotion, UUID> {

    Optional<Promotion> findByCode(String code);

    /**
     * Returns all promotions that are active and whose validity window
     * contains the supplied timestamp (typically Instant.now()).
     */
    List<Promotion> findByActiveTrueAndStartsAtBeforeAndExpiresAtAfter(
            Instant now1, Instant now2);
}
