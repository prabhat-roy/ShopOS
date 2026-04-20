package com.shopos.supplierportalservice.dto;

import com.shopos.supplierportalservice.domain.SupplierInvoice;
import com.shopos.supplierportalservice.domain.SupplierInvoiceStatus;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
public class InvoiceResponse {

    private UUID id;
    private UUID vendorId;
    private UUID purchaseOrderId;
    private String invoiceNumber;
    private BigDecimal amount;
    private String currency;
    private SupplierInvoiceStatus status;
    private LocalDate dueDate;
    private LocalDateTime paidAt;
    private String lineItems;
    private String notes;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    public static InvoiceResponse from(SupplierInvoice invoice) {
        return InvoiceResponse.builder()
                .id(invoice.getId())
                .vendorId(invoice.getVendorId())
                .purchaseOrderId(invoice.getPurchaseOrderId())
                .invoiceNumber(invoice.getInvoiceNumber())
                .amount(invoice.getAmount())
                .currency(invoice.getCurrency())
                .status(invoice.getStatus())
                .dueDate(invoice.getDueDate())
                .paidAt(invoice.getPaidAt())
                .lineItems(invoice.getLineItems())
                .notes(invoice.getNotes())
                .createdAt(invoice.getCreatedAt())
                .updatedAt(invoice.getUpdatedAt())
                .build();
    }
}
