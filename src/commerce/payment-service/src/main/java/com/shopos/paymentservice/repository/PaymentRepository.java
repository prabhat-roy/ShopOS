package com.shopos.paymentservice.repository;

import com.shopos.paymentservice.domain.Payment;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface PaymentRepository extends JpaRepository<Payment, UUID> {

    List<Payment> findByOrderId(String orderId);

    Page<Payment> findByCustomerId(String customerId, Pageable pageable);
}
