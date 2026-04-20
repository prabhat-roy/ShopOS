package com.shopos.supplierportalservice.repository;

import com.shopos.supplierportalservice.domain.SupplierInvoice;
import com.shopos.supplierportalservice.domain.SupplierInvoiceStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface SupplierInvoiceRepository extends JpaRepository<SupplierInvoice, UUID> {

    List<SupplierInvoice> findByVendorId(UUID vendorId);

    List<SupplierInvoice> findByStatus(SupplierInvoiceStatus status);

    Optional<SupplierInvoice> findByInvoiceNumber(String invoiceNumber);

    List<SupplierInvoice> findByVendorIdAndStatus(UUID vendorId, SupplierInvoiceStatus status);
}
