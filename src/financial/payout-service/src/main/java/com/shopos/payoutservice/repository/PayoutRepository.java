package com.shopos.payoutservice.repository;

import com.shopos.payoutservice.domain.Payout;
import com.shopos.payoutservice.domain.PayoutStatus;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Repository
public interface PayoutRepository extends JpaRepository<Payout, UUID> {

    Page<Payout> findByVendorId(UUID vendorId, Pageable pageable);

    List<Payout> findByStatus(PayoutStatus status);

    Page<Payout> findByVendorIdAndStatus(UUID vendorId, PayoutStatus status, Pageable pageable);

    /**
     * Returns all payouts whose scheduled time is before the given instant
     * and whose status matches the given value.
     * Used to identify PENDING payouts that are now due for processing.
     */
    List<Payout> findByScheduledAtBeforeAndStatus(LocalDateTime scheduledAt, PayoutStatus status);

    /**
     * Returns all PENDING payouts whose scheduledAt is in the past (or has no schedule set).
     * Payouts with no scheduledAt are treated as immediately due.
     */
    @Query("SELECT p FROM Payout p WHERE p.status = com.shopos.payoutservice.domain.PayoutStatus.PENDING " +
           "AND (p.scheduledAt IS NULL OR p.scheduledAt <= :now)")
    List<Payout> findDuePayouts(@Param("now") LocalDateTime now);
}
