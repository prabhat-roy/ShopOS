package com.shopos.supplierportalservice.service;

import com.shopos.supplierportalservice.domain.SupplierCatalogItem;
import com.shopos.supplierportalservice.domain.SupplierInvoice;
import com.shopos.supplierportalservice.domain.SupplierInvoiceStatus;
import com.shopos.supplierportalservice.dto.CreateInvoiceRequest;
import com.shopos.supplierportalservice.dto.UpsertCatalogItemRequest;
import com.shopos.supplierportalservice.repository.SupplierCatalogRepository;
import com.shopos.supplierportalservice.repository.SupplierInvoiceRepository;
import jakarta.persistence.EntityNotFoundException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class SupplierPortalService {

    private final SupplierInvoiceRepository invoiceRepository;
    private final SupplierCatalogRepository catalogRepository;

    // =========================================================================
    // Invoice operations
    // =========================================================================

    @Transactional
    public SupplierInvoice createInvoice(CreateInvoiceRequest req) {
        SupplierInvoice invoice = new SupplierInvoice(
                req.getVendorId(),
                req.getPurchaseOrderId(),
                req.getInvoiceNumber(),
                req.getAmount(),
                req.getCurrency(),
                req.getDueDate(),
                req.getLineItems(),
                req.getNotes());
        SupplierInvoice saved = invoiceRepository.save(invoice);
        log.info("Invoice created. id={} invoiceNumber={} vendorId={}",
                saved.getId(), saved.getInvoiceNumber(), saved.getVendorId());
        return saved;
    }

    @Transactional
    public SupplierInvoice submitInvoice(UUID invoiceId) {
        SupplierInvoice invoice = requireInvoice(invoiceId);
        requireStatus(invoice, SupplierInvoiceStatus.DRAFT);
        invoice.setStatus(SupplierInvoiceStatus.SUBMITTED);
        log.info("Invoice submitted. id={}", invoiceId);
        return invoiceRepository.save(invoice);
    }

    @Transactional
    public SupplierInvoice approveInvoice(UUID invoiceId) {
        SupplierInvoice invoice = requireInvoice(invoiceId);
        if (invoice.getStatus() != SupplierInvoiceStatus.SUBMITTED
                && invoice.getStatus() != SupplierInvoiceStatus.UNDER_REVIEW) {
            throw new IllegalStateException(
                    "Invoice can only be approved from SUBMITTED or UNDER_REVIEW status. "
                    + "Current status: " + invoice.getStatus());
        }
        invoice.setStatus(SupplierInvoiceStatus.APPROVED);
        log.info("Invoice approved. id={}", invoiceId);
        return invoiceRepository.save(invoice);
    }

    @Transactional
    public SupplierInvoice rejectInvoice(UUID invoiceId, String reason) {
        SupplierInvoice invoice = requireInvoice(invoiceId);
        if (invoice.getStatus() == SupplierInvoiceStatus.PAID) {
            throw new IllegalStateException("Cannot reject a PAID invoice.");
        }
        invoice.setStatus(SupplierInvoiceStatus.REJECTED);
        String existingNotes = invoice.getNotes() == null ? "" : invoice.getNotes();
        invoice.setNotes(existingNotes + "\nREJECTION REASON: " + reason);
        log.info("Invoice rejected. id={} reason={}", invoiceId, reason);
        return invoiceRepository.save(invoice);
    }

    @Transactional
    public SupplierInvoice markPaid(UUID invoiceId) {
        SupplierInvoice invoice = requireInvoice(invoiceId);
        requireStatus(invoice, SupplierInvoiceStatus.APPROVED);
        invoice.setStatus(SupplierInvoiceStatus.PAID);
        invoice.setPaidAt(LocalDateTime.now());
        log.info("Invoice marked as paid. id={}", invoiceId);
        return invoiceRepository.save(invoice);
    }

    @Transactional(readOnly = true)
    public SupplierInvoice getInvoice(UUID invoiceId) {
        return requireInvoice(invoiceId);
    }

    @Transactional(readOnly = true)
    public List<SupplierInvoice> listInvoices(UUID vendorId, SupplierInvoiceStatus status) {
        if (vendorId != null && status != null) {
            return invoiceRepository.findByVendorIdAndStatus(vendorId, status);
        }
        if (vendorId != null) {
            return invoiceRepository.findByVendorId(vendorId);
        }
        if (status != null) {
            return invoiceRepository.findByStatus(status);
        }
        return invoiceRepository.findAll();
    }

    // =========================================================================
    // Catalog item operations
    // =========================================================================

    @Transactional
    public SupplierCatalogItem upsertCatalogItem(UpsertCatalogItemRequest req) {
        Optional<SupplierCatalogItem> existing =
                catalogRepository.findByVendorIdAndSku(req.getVendorId(), req.getSku());

        SupplierCatalogItem item;
        if (existing.isPresent()) {
            item = existing.get();
            item.setProductId(req.getProductId());
            item.setProductName(req.getProductName());
            item.setUnitPrice(req.getUnitPrice());
            item.setCurrency(req.getCurrency() != null ? req.getCurrency() : "USD");
            item.setMinOrderQty(Math.max(1, req.getMinOrderQty()));
            item.setLeadTimeDays(req.getLeadTimeDays());
            item.setActive(true);
            log.info("Catalog item updated. vendorId={} sku={}", req.getVendorId(), req.getSku());
        } else {
            item = new SupplierCatalogItem(
                    req.getVendorId(),
                    req.getProductId(),
                    req.getSku(),
                    req.getProductName(),
                    req.getUnitPrice(),
                    req.getCurrency(),
                    req.getMinOrderQty(),
                    req.getLeadTimeDays());
            log.info("Catalog item created. vendorId={} sku={}", req.getVendorId(), req.getSku());
        }
        return catalogRepository.save(item);
    }

    @Transactional(readOnly = true)
    public SupplierCatalogItem getCatalogItem(UUID itemId) {
        return catalogRepository.findById(itemId)
                .orElseThrow(() -> new EntityNotFoundException(
                        "Catalog item not found: " + itemId));
    }

    @Transactional(readOnly = true)
    public List<SupplierCatalogItem> listCatalogItems(UUID vendorId, boolean activeOnly) {
        if (vendorId != null && activeOnly) {
            return catalogRepository.findByVendorIdAndActive(vendorId, true);
        }
        if (vendorId != null) {
            return catalogRepository.findByVendorId(vendorId);
        }
        return catalogRepository.findAll();
    }

    @Transactional
    public SupplierCatalogItem deactivateCatalogItem(UUID itemId) {
        SupplierCatalogItem item = getCatalogItem(itemId);
        item.setActive(false);
        log.info("Catalog item deactivated. id={}", itemId);
        return catalogRepository.save(item);
    }

    // =========================================================================
    // Private helpers
    // =========================================================================

    private SupplierInvoice requireInvoice(UUID invoiceId) {
        return invoiceRepository.findById(invoiceId)
                .orElseThrow(() -> new EntityNotFoundException(
                        "Invoice not found: " + invoiceId));
    }

    private void requireStatus(SupplierInvoice invoice, SupplierInvoiceStatus expected) {
        if (invoice.getStatus() != expected) {
            throw new IllegalStateException(
                    "Expected invoice status " + expected
                    + " but was " + invoice.getStatus() + ". Invoice id: " + invoice.getId());
        }
    }
}
