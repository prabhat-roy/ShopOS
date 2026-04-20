package com.shopos.invoiceservice.service;

import com.shopos.invoiceservice.domain.Invoice;
import com.shopos.invoiceservice.domain.InvoiceStatus;
import com.shopos.invoiceservice.dto.CreateInvoiceRequest;
import com.shopos.invoiceservice.dto.InvoiceResponse;
import com.shopos.invoiceservice.repository.InvoiceRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.YearMonth;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class InvoiceService {

    private static final DateTimeFormatter YEAR_MONTH_FMT = DateTimeFormatter.ofPattern("yyyyMM");

    private final InvoiceRepository invoiceRepository;

    /**
     * Creates a new invoice in DRAFT status and generates its invoice number.
     * Invoice number format: INV-YYYYMM-XXXXX where XXXXX is the first 5 uppercase hex chars of a UUID.
     */
    @Transactional
    public InvoiceResponse createInvoice(CreateInvoiceRequest request) {
        String invoiceNumber = generateInvoiceNumber();

        Invoice invoice = Invoice.builder()
                .orderId(request.orderId())
                .customerId(request.customerId())
                .invoiceNumber(invoiceNumber)
                .status(InvoiceStatus.DRAFT)
                .subtotal(request.subtotal())
                .taxAmount(request.taxAmount())
                .totalAmount(request.totalAmount())
                .currency(request.currency() != null ? request.currency() : "USD")
                .lineItems(request.lineItems())
                .billingAddress(request.billingAddress())
                .dueDate(request.dueDate())
                .notes(request.notes())
                .build();

        Invoice saved = invoiceRepository.save(invoice);
        log.info("Created invoice {} for orderId={} customerId={}", saved.getInvoiceNumber(),
                saved.getOrderId(), saved.getCustomerId());
        return InvoiceResponse.from(saved);
    }

    /**
     * Retrieves a single invoice by its UUID.
     *
     * @throws jakarta.persistence.EntityNotFoundException if no invoice exists with that id
     */
    @Transactional(readOnly = true)
    public InvoiceResponse getInvoice(UUID id) {
        Invoice invoice = findOrThrow(id);
        return InvoiceResponse.from(invoice);
    }

    /**
     * Returns a paginated list of invoices for a given customer, optionally filtered by status.
     */
    @Transactional(readOnly = true)
    public Page<InvoiceResponse> listByCustomer(UUID customerId, InvoiceStatus status, Pageable pageable) {
        Page<Invoice> page;
        if (status != null) {
            page = invoiceRepository.findByCustomerIdAndStatus(customerId, status, pageable);
        } else {
            page = invoiceRepository.findByCustomerId(customerId, pageable);
        }
        return page.map(InvoiceResponse::from);
    }

    /**
     * Returns all invoices with the given status.
     */
    @Transactional(readOnly = true)
    public List<InvoiceResponse> listByStatus(InvoiceStatus status) {
        return invoiceRepository.findByStatus(status)
                .stream()
                .map(InvoiceResponse::from)
                .toList();
    }

    /**
     * Transitions a DRAFT invoice to ISSUED.
     *
     * @throws IllegalStateException if the invoice is not in DRAFT status
     */
    @Transactional
    public void issueInvoice(UUID id) {
        Invoice invoice = findOrThrow(id);
        requireStatus(invoice, InvoiceStatus.DRAFT, "issue");
        invoice.setStatus(InvoiceStatus.ISSUED);
        invoiceRepository.save(invoice);
        log.info("Issued invoice {}", invoice.getInvoiceNumber());
    }

    /**
     * Transitions an ISSUED invoice to SENT.
     *
     * @throws IllegalStateException if the invoice is not in ISSUED status
     */
    @Transactional
    public void markSent(UUID id) {
        Invoice invoice = findOrThrow(id);
        requireStatus(invoice, InvoiceStatus.ISSUED, "mark as sent");
        invoice.setStatus(InvoiceStatus.SENT);
        invoiceRepository.save(invoice);
        log.info("Marked invoice {} as SENT", invoice.getInvoiceNumber());
    }

    /**
     * Marks an invoice as PAID and records the payment timestamp.
     * Accepts invoices in ISSUED, SENT, or OVERDUE status.
     *
     * @throws IllegalStateException if the invoice is already PAID, CANCELLED, or VOID
     */
    @Transactional
    public void markPaid(UUID id) {
        Invoice invoice = findOrThrow(id);
        if (invoice.getStatus() == InvoiceStatus.PAID ||
            invoice.getStatus() == InvoiceStatus.CANCELLED ||
            invoice.getStatus() == InvoiceStatus.VOID) {
            throw new IllegalStateException(
                    "Cannot mark invoice " + invoice.getInvoiceNumber() +
                    " as PAID because it is already " + invoice.getStatus());
        }
        invoice.setStatus(InvoiceStatus.PAID);
        invoice.setPaidAt(LocalDateTime.now());
        invoiceRepository.save(invoice);
        log.info("Marked invoice {} as PAID", invoice.getInvoiceNumber());
    }

    /**
     * Manually marks an ISSUED or SENT invoice as OVERDUE.
     *
     * @throws IllegalStateException if the invoice is not in ISSUED or SENT status
     */
    @Transactional
    public void markOverdue(UUID id) {
        Invoice invoice = findOrThrow(id);
        if (invoice.getStatus() != InvoiceStatus.ISSUED &&
            invoice.getStatus() != InvoiceStatus.SENT) {
            throw new IllegalStateException(
                    "Cannot mark invoice " + invoice.getInvoiceNumber() +
                    " as OVERDUE. Current status: " + invoice.getStatus());
        }
        invoice.setStatus(InvoiceStatus.OVERDUE);
        invoiceRepository.save(invoice);
        log.info("Marked invoice {} as OVERDUE", invoice.getInvoiceNumber());
    }

    /**
     * Cancels an invoice that has not yet been paid.
     *
     * @throws IllegalStateException if the invoice is in a terminal state (PAID, VOID)
     */
    @Transactional
    public void cancelInvoice(UUID id) {
        Invoice invoice = findOrThrow(id);
        if (invoice.getStatus() == InvoiceStatus.PAID ||
            invoice.getStatus() == InvoiceStatus.VOID) {
            throw new IllegalStateException(
                    "Cannot cancel invoice " + invoice.getInvoiceNumber() +
                    " because it is already " + invoice.getStatus());
        }
        invoice.setStatus(InvoiceStatus.CANCELLED);
        invoiceRepository.save(invoice);
        log.info("Cancelled invoice {}", invoice.getInvoiceNumber());
    }

    /**
     * Voids an issued invoice (legal nullification after issuance).
     * Allowed from any status except VOID.
     *
     * @throws IllegalStateException if the invoice is already VOID
     */
    @Transactional
    public void voidInvoice(UUID id) {
        Invoice invoice = findOrThrow(id);
        if (invoice.getStatus() == InvoiceStatus.VOID) {
            throw new IllegalStateException(
                    "Invoice " + invoice.getInvoiceNumber() + " is already VOID");
        }
        invoice.setStatus(InvoiceStatus.VOID);
        invoiceRepository.save(invoice);
        log.info("Voided invoice {}", invoice.getInvoiceNumber());
    }

    /**
     * Batch operation: marks all ISSUED and SENT invoices whose due date is before today as OVERDUE.
     * Uses a single bulk UPDATE query for efficiency.
     *
     * @return the number of invoices updated
     */
    @Transactional
    public int detectOverdue() {
        int count = invoiceRepository.bulkMarkOverdue(LocalDate.now());
        log.info("detectOverdue: marked {} invoice(s) as OVERDUE", count);
        return count;
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    private Invoice findOrThrow(UUID id) {
        return invoiceRepository.findById(id)
                .orElseThrow(() -> new jakarta.persistence.EntityNotFoundException(
                        "Invoice not found with id: " + id));
    }

    private void requireStatus(Invoice invoice, InvoiceStatus required, String operation) {
        if (invoice.getStatus() != required) {
            throw new IllegalStateException(
                    "Cannot " + operation + " invoice " + invoice.getInvoiceNumber() +
                    ". Required status: " + required + ", actual: " + invoice.getStatus());
        }
    }

    /**
     * Generates a unique invoice number in the format INV-YYYYMM-XXXXX.
     * XXXXX = first 5 uppercase hex characters of a random UUID.
     */
    private String generateInvoiceNumber() {
        String yearMonth = YearMonth.now().format(YEAR_MONTH_FMT);
        String suffix = UUID.randomUUID().toString().replace("-", "").substring(0, 5).toUpperCase();
        return "INV-" + yearMonth + "-" + suffix;
    }
}
