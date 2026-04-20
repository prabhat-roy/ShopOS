package com.shopos.vendorservice.repository;

import com.shopos.vendorservice.domain.Vendor;
import com.shopos.vendorservice.domain.VendorStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface VendorRepository extends JpaRepository<Vendor, UUID> {

    List<Vendor> findByStatus(VendorStatus status);

    Optional<Vendor> findByEmail(String email);

    List<Vendor> findByCountry(String country);

    boolean existsByEmail(String email);
}
