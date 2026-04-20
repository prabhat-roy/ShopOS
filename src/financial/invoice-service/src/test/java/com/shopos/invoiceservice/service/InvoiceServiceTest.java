package com.shopos.invoiceservice.service;

import com.shopos.invoiceservice.domain.Invoice;
import com.shopos.invoiceservice.domain.InvoiceStatus;
import com.shopos.invoiceservice.dto.CreateInvoiceRequest;
import com.shopos.invoiceservice.dto.InvoiceResponse;
import com.shopos.invoiceservice.repository.InvoiceRepository;
import jakarta.persistence.EntityNotFoundException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@DisplayName("InvoiceService unit tests")
class InvoiceServiceTest {

    @Mock
    private InvoiceRepository invoiceRepository;

    @InjectMocks
    private InvoiceService invoiceService;

    private UUID invoiceId;
    private Invoice draftInvoice;

    @BeforeEach
    void setUp() {
        invoiceId = UUID.randomUUID();
        draftInvoice = Invoice.builder()
                .id(invoiceId)
                .orderId(UUID.randomUUID())
                .customerId(UUID.randomUUID())
                .invoiceNumber("INV-202601-ABCDE")
                .status(InvoiceStatus.DRAFT)
                .subtotal(new BigDecimal("100.00"))
                .taxAmount(new BigDecimal("10.00"))
                .totalAmount(new BigDecimal("110.00"))
                .currency("USD")
                .lineItems("[{\"sku\":\"SKU-001\",\"qty\":1,\"price\":100.00}]")
                .billingAddress("123 Main St, Springfield")
                .dueDate(LocalDate.now().plusDays(30))
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();
    }

    // -----------------------------------------------------------------------
    // Test 1 — createInvoice generates a properly formatted invoice number
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("createInvoice — should generate INV-YYYYMM-XXXXX invoice number")
    void createInvoice_generatesCorrectInvoiceNumber() {
        CreateInvoiceRequest request = buildCreateRequest();
        when(invoiceRepository.save(any(Invoice.class))).thenAnswer(inv -> inv.getArgument(0));

        InvoiceResponse response = invoiceService.createInvoice(request);

        assertThat(response.invoiceNumber()).matches("INV-\\d{6}-[A-F0-9]{5}");
    }

    // -----------------------------------------------------------------------
    // Test 2 — createInvoice persists DRAFT status
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("createInvoice — new invoice must be in DRAFT status")
    void createInvoice_statusIsDraft() {
        CreateInvoiceRequest request = buildCreateRequest();
        when(invoiceRepository.save(any(Invoice.class))).thenAnswer(inv -> inv.getArgument(0));

        InvoiceResponse response = invoiceService.createInvoice(request);

        assertThat(response.status()).isEqualTo(InvoiceStatus.DRAFT);
    }

    // -----------------------------------------------------------------------
    // Test 3 — getInvoice throws EntityNotFoundException for unknown id
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("getInvoice — throws EntityNotFoundException when invoice not found")
    void getInvoice_throwsWhenNotFound() {
        UUID unknownId = UUID.randomUUID();
        when(invoiceRepository.findById(unknownId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> invoiceService.getInvoice(unknownId))
                .isInstanceOf(EntityNotFoundException.class)
                .hasMessageContaining(unknownId.toString());
    }

    // -----------------------------------------------------------------------
    // Test 4 — issueInvoice transitions DRAFT → ISSUED
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("issueInvoice — transitions DRAFT to ISSUED")
    void issueInvoice_draftToIssued() {
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(draftInvoice));
        when(invoiceRepository.save(any(Invoice.class))).thenAnswer(inv -> inv.getArgument(0));

        invoiceService.issueInvoice(invoiceId);

        ArgumentCaptor<Invoice> captor = ArgumentCaptor.forClass(Invoice.class);
        verify(invoiceRepository).save(captor.capture());
        assertThat(captor.getValue().getStatus()).isEqualTo(InvoiceStatus.ISSUED);
    }

    // -----------------------------------------------------------------------
    // Test 5 — issueInvoice throws when invoice is not DRAFT
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("issueInvoice — throws IllegalStateException when invoice is not DRAFT")
    void issueInvoice_throwsWhenNotDraft() {
        draftInvoice.setStatus(InvoiceStatus.ISSUED);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(draftInvoice));

        assertThatThrownBy(() -> invoiceService.issueInvoice(invoiceId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("DRAFT");
    }

    // -----------------------------------------------------------------------
    // Test 6 — markPaid sets paidAt timestamp
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("markPaid — sets paidAt and transitions to PAID")
    void markPaid_setsPaidAt() {
        draftInvoice.setStatus(InvoiceStatus.SENT);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(draftInvoice));
        when(invoiceRepository.save(any(Invoice.class))).thenAnswer(inv -> inv.getArgument(0));

        invoiceService.markPaid(invoiceId);

        ArgumentCaptor<Invoice> captor = ArgumentCaptor.forClass(Invoice.class);
        verify(invoiceRepository).save(captor.capture());
        Invoice saved = captor.getValue();
        assertThat(saved.getStatus()).isEqualTo(InvoiceStatus.PAID);
        assertThat(saved.getPaidAt()).isNotNull();
    }

    // -----------------------------------------------------------------------
    // Test 7 — markPaid throws when invoice is already CANCELLED
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("markPaid — throws IllegalStateException when invoice is CANCELLED")
    void markPaid_throwsWhenCancelled() {
        draftInvoice.setStatus(InvoiceStatus.CANCELLED);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(draftInvoice));

        assertThatThrownBy(() -> invoiceService.markPaid(invoiceId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("CANCELLED");
    }

    // -----------------------------------------------------------------------
    // Test 8 — cancelInvoice transitions ISSUED → CANCELLED
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("cancelInvoice — transitions ISSUED to CANCELLED")
    void cancelInvoice_issuedToCancelled() {
        draftInvoice.setStatus(InvoiceStatus.ISSUED);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(draftInvoice));
        when(invoiceRepository.save(any(Invoice.class))).thenAnswer(inv -> inv.getArgument(0));

        invoiceService.cancelInvoice(invoiceId);

        ArgumentCaptor<Invoice> captor = ArgumentCaptor.forClass(Invoice.class);
        verify(invoiceRepository).save(captor.capture());
        assertThat(captor.getValue().getStatus()).isEqualTo(InvoiceStatus.CANCELLED);
    }

    // -----------------------------------------------------------------------
    // Test 9 — cancelInvoice throws when invoice is already PAID
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("cancelInvoice — throws IllegalStateException when invoice is PAID")
    void cancelInvoice_throwsWhenPaid() {
        draftInvoice.setStatus(InvoiceStatus.PAID);
        when(invoiceRepository.findById(invoiceId)).thenReturn(Optional.of(draftInvoice));

        assertThatThrownBy(() -> invoiceService.cancelInvoice(invoiceId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("PAID");
    }

    // -----------------------------------------------------------------------
    // Test 10 — detectOverdue calls bulkMarkOverdue and returns count
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("detectOverdue — delegates to repository and returns the updated count")
    void detectOverdue_returnsBulkUpdateCount() {
        when(invoiceRepository.bulkMarkOverdue(any(LocalDate.class))).thenReturn(5);

        int count = invoiceService.detectOverdue();

        assertThat(count).isEqualTo(5);
        verify(invoiceRepository).bulkMarkOverdue(LocalDate.now());
    }

    // -----------------------------------------------------------------------
    // Helper
    // -----------------------------------------------------------------------
    private CreateInvoiceRequest buildCreateRequest() {
        return new CreateInvoiceRequest(
                UUID.randomUUID(),
                UUID.randomUUID(),
                new BigDecimal("100.00"),
                new BigDecimal("10.00"),
                new BigDecimal("110.00"),
                "USD",
                "[{\"sku\":\"SKU-001\",\"qty\":1,\"price\":100.00}]",
                "123 Main St, Springfield",
                LocalDate.now().plusDays(30),
                "Test invoice"
        );
    }
}
