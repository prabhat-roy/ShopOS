package com.shopos.supplierportalservice.repository;

import com.shopos.supplierportalservice.domain.SupplierCatalogItem;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface SupplierCatalogRepository extends JpaRepository<SupplierCatalogItem, UUID> {

    List<SupplierCatalogItem> findByVendorId(UUID vendorId);

    List<SupplierCatalogItem> findByVendorIdAndActive(UUID vendorId, boolean active);

    Optional<SupplierCatalogItem> findByVendorIdAndSku(UUID vendorId, String sku);
}
