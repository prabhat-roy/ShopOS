package com.shopos.supplierportalservice.service;

import com.shopos.supplierportalservice.domain.SupplierCatalogItem;
import com.shopos.supplierportalservice.domain.SupplierInvoice;
import com.shopos.supplierportalservice.domain.SupplierInvoiceStatus;
import com.shopos.supplierportalservice.dto.CreateInvoiceRequest;
import com.shopos.supplierportalservice.dto.UpsertCatalogItemRequest;
import com.shopos.supplierportalservice.repository.SupplierCatalogRepository;
import com.shopos.supplierportalservice.repository.SupplierInvoiceRepository;
import jakarta.persistence.EntityNotFoundException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class SupplierPortalServiceTest {

    @Mock
    private SupplierInvoiceRepository invoiceRepository;

    @Mock
    private SupplierCatalogRepository catalogRepository;

    @InjectMocks
    private SupplierPortalService service;

    private UUID vendorId;
    private UUID invoiceId;

    @BeforeEach
    void setUp() {
        vendorId = UUID.randomUUID();
        invoiceId = UUID.randomUUID();
    }

    // ------------------------------------------------------------------
    // Helper to build a saved invoice stub
    // ------------------------------------------------------------------

    private SupplierInvoice buildInvoice(SupplierInvoiceStatus status) {
        SupplierInvoice inv = new SupplierInvoice(
                vendorId, null, "INV-001",
                BigDecimal.valueOf(500.00), "USD",
                LocalDate.now().plusDays(30), null, null);
        // Use reflection-free: set ID via spy is avoided; we use a real instance
        // and rely on mocked repo returning it with the desired state.
        inv.setStatus(status);
        return inv;
    }

    // ------------------------------------------------------------------
    // 1. createInvoice saves and returns a DRAFT invoice
    // ------------------------------------------------------------------
    @Test
    void createInvoice_savesDraftInvoice() {
        CreateInvoiceRequest req = new CreateInvoiceRequest();
        req.setVendorId(vendorId);
        req.setInvoiceNumber("INV-001");
        req.setAmount(BigDecimal.valueOf(999.99));
        req.setCurrency("USD");
        req.setDueDate(LocalDate.now().plusDays(30));

        SupplierInvoice saved = buildInvoice(SupplierInvoiceStatus.DRAFT);
        when(invoiceRepository.save(any())).thenReturn(saved);

        SupplierInvoice result = service.createInvoice(req);

        assertThat(result.getStatus()).isEqualTo(SupplierInvoiceStatus.DRAFT);
        verify(invoiceRepository).save(any(SupplierInvoice.class));
    }

    // ------------------------------------------------------------------
    // 2. submitInvoice transitions DRAFT → SUBMITTED
    // ------------------------------------------------------------------
    @Test
    void submitInvoice_transitionsDraftToSubmitted() {
        SupplierInvoice draft = buildInvoice(SupplierInvoiceStatus.DRAFT);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(draft));

        SupplierInvoice submitted = buildInvoice(SupplierInvoiceStatus.SUBMITTED);
        when(invoiceRepository.save(draft)).thenReturn(submitted);

        SupplierInvoice result = service.submitInvoice(invoiceId);

        assertThat(result.getStatus()).isEqualTo(SupplierInvoiceStatus.SUBMITTED);
    }

    // ------------------------------------------------------------------
    // 3. submitInvoice throws when invoice is not DRAFT
    // ------------------------------------------------------------------
    @Test
    void submitInvoice_throwsWhenNotDraft() {
        SupplierInvoice alreadySubmitted = buildInvoice(SupplierInvoiceStatus.SUBMITTED);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(alreadySubmitted));

        assertThatThrownBy(() -> service.submitInvoice(invoiceId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("DRAFT");
    }

    // ------------------------------------------------------------------
    // 4. approveInvoice transitions SUBMITTED → APPROVED
    // ------------------------------------------------------------------
    @Test
    void approveInvoice_transitionsSubmittedToApproved() {
        SupplierInvoice submitted = buildInvoice(SupplierInvoiceStatus.SUBMITTED);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(submitted));

        SupplierInvoice approved = buildInvoice(SupplierInvoiceStatus.APPROVED);
        when(invoiceRepository.save(submitted)).thenReturn(approved);

        SupplierInvoice result = service.approveInvoice(invoiceId);

        assertThat(result.getStatus()).isEqualTo(SupplierInvoiceStatus.APPROVED);
    }

    // ------------------------------------------------------------------
    // 5. rejectInvoice appends reason to notes
    // ------------------------------------------------------------------
    @Test
    void rejectInvoice_appendsReasonToNotes() {
        SupplierInvoice submitted = buildInvoice(SupplierInvoiceStatus.SUBMITTED);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(submitted));
        when(invoiceRepository.save(submitted)).thenReturn(submitted);

        service.rejectInvoice(invoiceId, "Duplicate invoice");

        ArgumentCaptor<SupplierInvoice> captor = ArgumentCaptor.forClass(SupplierInvoice.class);
        verify(invoiceRepository).save(captor.capture());

        SupplierInvoice saved = captor.getValue();
        assertThat(saved.getStatus()).isEqualTo(SupplierInvoiceStatus.REJECTED);
        assertThat(saved.getNotes()).contains("Duplicate invoice");
    }

    // ------------------------------------------------------------------
    // 6. markPaid sets APPROVED → PAID and populates paidAt
    // ------------------------------------------------------------------
    @Test
    void markPaid_setsStatusAndPaidAt() {
        SupplierInvoice approved = buildInvoice(SupplierInvoiceStatus.APPROVED);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(approved));
        when(invoiceRepository.save(approved)).thenReturn(approved);

        service.markPaid(invoiceId);

        ArgumentCaptor<SupplierInvoice> captor = ArgumentCaptor.forClass(SupplierInvoice.class);
        verify(invoiceRepository).save(captor.capture());

        SupplierInvoice saved = captor.getValue();
        assertThat(saved.getStatus()).isEqualTo(SupplierInvoiceStatus.PAID);
        assertThat(saved.getPaidAt()).isNotNull();
    }

    // ------------------------------------------------------------------
    // 7. getInvoice throws EntityNotFoundException when not found
    // ------------------------------------------------------------------
    @Test
    void getInvoice_throwsWhenNotFound() {
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> service.getInvoice(invoiceId))
                .isInstanceOf(EntityNotFoundException.class)
                .hasMessageContaining(invoiceId.toString());
    }

    // ------------------------------------------------------------------
    // 8. upsertCatalogItem creates new item when none exists
    // ------------------------------------------------------------------
    @Test
    void upsertCatalogItem_createsNewItemWhenAbsent() {
        UpsertCatalogItemRequest req = new UpsertCatalogItemRequest();
        req.setVendorId(vendorId);
        req.setProductId("prod-123");
        req.setSku("SKU-001");
        req.setProductName("Widget Pro");
        req.setUnitPrice(BigDecimal.valueOf(29.99));
        req.setCurrency("USD");
        req.setMinOrderQty(5);
        req.setLeadTimeDays(7);

        when(catalogRepository.findByVendorIdAndSku(vendorId, "SKU-001"))
                .thenReturn(Optional.empty());

        SupplierCatalogItem created = new SupplierCatalogItem(
                vendorId, "prod-123", "SKU-001", "Widget Pro",
                BigDecimal.valueOf(29.99), "USD", 5, 7);
        when(catalogRepository.save(any())).thenReturn(created);

        SupplierCatalogItem result = service.upsertCatalogItem(req);

        assertThat(result.getSku()).isEqualTo("SKU-001");
        assertThat(result.isActive()).isTrue();
        verify(catalogRepository).save(any(SupplierCatalogItem.class));
    }

    // ------------------------------------------------------------------
    // 9. upsertCatalogItem updates existing item when found
    // ------------------------------------------------------------------
    @Test
    void upsertCatalogItem_updatesExistingItem() {
        SupplierCatalogItem existing = new SupplierCatalogItem(
                vendorId, "prod-old", "SKU-001", "Old Name",
                BigDecimal.valueOf(10.00), "USD", 1, 3);
        when(catalogRepository.findByVendorIdAndSku(vendorId, "SKU-001"))
                .thenReturn(Optional.of(existing));
        when(catalogRepository.save(existing)).thenReturn(existing);

        UpsertCatalogItemRequest req = new UpsertCatalogItemRequest();
        req.setVendorId(vendorId);
        req.setProductId("prod-new");
        req.setSku("SKU-001");
        req.setProductName("New Name");
        req.setUnitPrice(BigDecimal.valueOf(25.00));
        req.setCurrency("USD");
        req.setMinOrderQty(2);
        req.setLeadTimeDays(5);

        service.upsertCatalogItem(req);

        ArgumentCaptor<SupplierCatalogItem> captor = ArgumentCaptor.forClass(SupplierCatalogItem.class);
        verify(catalogRepository).save(captor.capture());
        SupplierCatalogItem updated = captor.getValue();
        assertThat(updated.getProductName()).isEqualTo("New Name");
        assertThat(updated.getUnitPrice()).isEqualByComparingTo(BigDecimal.valueOf(25.00));
    }

    // ------------------------------------------------------------------
    // 10. deactivateCatalogItem sets active=false
    // ------------------------------------------------------------------
    @Test
    void deactivateCatalogItem_setsActiveFalse() {
        UUID itemId = UUID.randomUUID();
        SupplierCatalogItem item = new SupplierCatalogItem(
                vendorId, "prod-1", "SKU-X", "Some Product",
                BigDecimal.valueOf(15.00), "USD", 1, 2);
        when(catalogRepository.findById(itemId)).thenReturn(Optional.of(item));
        when(catalogRepository.save(item)).thenReturn(item);

        service.deactivateCatalogItem(itemId);

        ArgumentCaptor<SupplierCatalogItem> captor = ArgumentCaptor.forClass(SupplierCatalogItem.class);
        verify(catalogRepository).save(captor.capture());
        assertThat(captor.getValue().isActive()).isFalse();
    }
}
