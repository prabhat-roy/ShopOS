package com.shopos.kycamlservice.repository;

import com.shopos.kycamlservice.domain.KycRecord;
import com.shopos.kycamlservice.domain.KycStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface KycRecordRepository extends JpaRepository<KycRecord, UUID> {

    Optional<KycRecord> findByCustomerId(UUID customerId);

    List<KycRecord> findByStatus(KycStatus status);

    /**
     * Used by the expiry detection batch job to find VERIFIED records
     * whose expiresAt timestamp is in the past.
     */
    List<KycRecord> findByExpiresAtBeforeAndStatus(LocalDateTime threshold, KycStatus status);

    boolean existsByCustomerId(UUID customerId);
}
