package com.shopos.supplierportalservice.domain;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(
    name = "supplier_invoices",
    uniqueConstraints = @UniqueConstraint(name = "uq_invoice_number", columnNames = "invoice_number")
)
@Getter
@Setter
@NoArgsConstructor
public class SupplierInvoice {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "vendor_id", nullable = false)
    private UUID vendorId;

    @Column(name = "purchase_order_id")
    private UUID purchaseOrderId;

    @Column(name = "invoice_number", nullable = false, unique = true, length = 100)
    private String invoiceNumber;

    @Column(name = "amount", nullable = false, precision = 19, scale = 4)
    private BigDecimal amount;

    @Column(name = "currency", nullable = false, length = 3)
    private String currency = "USD";

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false, length = 20)
    private SupplierInvoiceStatus status = SupplierInvoiceStatus.DRAFT;

    @Column(name = "due_date")
    private LocalDate dueDate;

    @Column(name = "paid_at")
    private LocalDateTime paidAt;

    @Column(name = "line_items", columnDefinition = "TEXT")
    private String lineItems;

    @Column(name = "notes", columnDefinition = "TEXT")
    private String notes;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;

    public SupplierInvoice(
            UUID vendorId,
            UUID purchaseOrderId,
            String invoiceNumber,
            BigDecimal amount,
            String currency,
            LocalDate dueDate,
            String lineItems,
            String notes) {
        this.vendorId = vendorId;
        this.purchaseOrderId = purchaseOrderId;
        this.invoiceNumber = invoiceNumber;
        this.amount = amount;
        this.currency = currency != null ? currency : "USD";
        this.dueDate = dueDate;
        this.lineItems = lineItems;
        this.notes = notes;
        this.status = SupplierInvoiceStatus.DRAFT;
    }
}
