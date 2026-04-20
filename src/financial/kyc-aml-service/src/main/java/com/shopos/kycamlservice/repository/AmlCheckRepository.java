package com.shopos.kycamlservice.repository;

import com.shopos.kycamlservice.domain.AmlCheck;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface AmlCheckRepository extends JpaRepository<AmlCheck, UUID> {

    List<AmlCheck> findByCustomerId(UUID customerId);

    List<AmlCheck> findByCheckTypeAndResult(String checkType, String result);

    List<AmlCheck> findByCustomerIdAndCheckType(UUID customerId, String checkType);
}
