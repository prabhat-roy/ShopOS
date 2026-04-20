package com.shopos.chargeback.repository;

import com.shopos.chargeback.domain.Chargeback;
import com.shopos.chargeback.domain.ChargebackStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface ChargebackRepository extends JpaRepository<Chargeback, UUID> {

    List<Chargeback> findByCustomerId(String customerId);

    List<Chargeback> findByPaymentId(String paymentId);

    List<Chargeback> findByStatus(ChargebackStatus status);
}
