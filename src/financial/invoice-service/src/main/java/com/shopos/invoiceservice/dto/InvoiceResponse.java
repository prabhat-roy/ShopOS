package com.shopos.invoiceservice.dto;

import com.shopos.invoiceservice.domain.Invoice;
import com.shopos.invoiceservice.domain.InvoiceStatus;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

/**
 * Outbound representation of an Invoice resource.
 */
public record InvoiceResponse(
        UUID id,
        UUID orderId,
        UUID customerId,
        String invoiceNumber,
        InvoiceStatus status,
        BigDecimal subtotal,
        BigDecimal taxAmount,
        BigDecimal totalAmount,
        String currency,
        String lineItems,
        String billingAddress,
        LocalDate dueDate,
        LocalDateTime paidAt,
        String notes,
        LocalDateTime createdAt,
        LocalDateTime updatedAt
) {

    /**
     * Factory method that maps an {@link Invoice} entity to an {@link InvoiceResponse} DTO.
     *
     * @param invoice the entity to convert
     * @return an immutable response record
     */
    public static InvoiceResponse from(Invoice invoice) {
        return new InvoiceResponse(
                invoice.getId(),
                invoice.getOrderId(),
                invoice.getCustomerId(),
                invoice.getInvoiceNumber(),
                invoice.getStatus(),
                invoice.getSubtotal(),
                invoice.getTaxAmount(),
                invoice.getTotalAmount(),
                invoice.getCurrency(),
                invoice.getLineItems(),
                invoice.getBillingAddress(),
                invoice.getDueDate(),
                invoice.getPaidAt(),
                invoice.getNotes(),
                invoice.getCreatedAt(),
                invoice.getUpdatedAt()
        );
    }
}
