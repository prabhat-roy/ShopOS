package com.shopos.invoiceservice.controller;

import com.shopos.invoiceservice.domain.InvoiceStatus;
import com.shopos.invoiceservice.dto.CreateInvoiceRequest;
import com.shopos.invoiceservice.dto.InvoiceResponse;
import com.shopos.invoiceservice.service.InvoiceService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.web.PageableDefault;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/invoices")
@RequiredArgsConstructor
public class InvoiceController {

    private final InvoiceService invoiceService;

    /**
     * POST /invoices — creates a new invoice in DRAFT status.
     * Returns 201 Created with the created resource.
     */
    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public InvoiceResponse createInvoice(@Valid @RequestBody CreateInvoiceRequest request) {
        return invoiceService.createInvoice(request);
    }

    /**
     * GET /invoices/{id} — retrieves a single invoice by UUID.
     */
    @GetMapping("/{id}")
    public InvoiceResponse getInvoice(@PathVariable UUID id) {
        return invoiceService.getInvoice(id);
    }

    /**
     * GET /invoices?customerId=&status=&page=&size=&sort=
     * Returns a paginated list filtered by customerId (required) and optional status.
     * Falls back to listing all invoices with a given status when customerId is absent.
     */
    @GetMapping
    public ResponseEntity<?> listInvoices(
            @RequestParam(required = false) UUID customerId,
            @RequestParam(required = false) InvoiceStatus status,
            @PageableDefault(size = 20, sort = "createdAt") Pageable pageable) {

        if (customerId != null) {
            Page<InvoiceResponse> page = invoiceService.listByCustomer(customerId, status, pageable);
            return ResponseEntity.ok(page);
        }
        if (status != null) {
            List<InvoiceResponse> list = invoiceService.listByStatus(status);
            return ResponseEntity.ok(list);
        }
        return ResponseEntity.badRequest()
                .body(Map.of("error", "Provide at least one of: customerId, status"));
    }

    /**
     * POST /invoices/{id}/issue — transitions DRAFT → ISSUED.
     */
    @PostMapping("/{id}/issue")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void issueInvoice(@PathVariable UUID id) {
        invoiceService.issueInvoice(id);
    }

    /**
     * POST /invoices/{id}/send — transitions ISSUED → SENT.
     */
    @PostMapping("/{id}/send")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void sendInvoice(@PathVariable UUID id) {
        invoiceService.markSent(id);
    }

    /**
     * POST /invoices/{id}/pay — records payment, transitions → PAID.
     */
    @PostMapping("/{id}/pay")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void payInvoice(@PathVariable UUID id) {
        invoiceService.markPaid(id);
    }

    /**
     * POST /invoices/{id}/overdue — manually marks ISSUED/SENT → OVERDUE.
     */
    @PostMapping("/{id}/overdue")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void markOverdue(@PathVariable UUID id) {
        invoiceService.markOverdue(id);
    }

    /**
     * DELETE /invoices/{id} — cancels the invoice (logical delete, sets status=CANCELLED).
     */
    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void cancelInvoice(@PathVariable UUID id) {
        invoiceService.cancelInvoice(id);
    }

    /**
     * POST /invoices/detect-overdue — batch job endpoint.
     * Marks all ISSUED/SENT invoices past their due date as OVERDUE.
     * Returns the count of records updated.
     */
    @PostMapping("/detect-overdue")
    public ResponseEntity<Map<String, Integer>> detectOverdue() {
        int count = invoiceService.detectOverdue();
        return ResponseEntity.ok(Map.of("updatedCount", count));
    }
}
