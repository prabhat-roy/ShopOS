package com.shopos.ediservice.dto;

import com.shopos.ediservice.domain.EdiFormat;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

/**
 * Request payload for the EDI generate endpoint.
 *
 * @param format       Target EDI format (X12 or EDIFACT).
 * @param documentType Business document type: PO, INVOICE, ASN, ACK.
 * @param senderId     EDI interchange sender ID (ISA06).
 * @param receiverId   EDI interchange receiver ID (ISA08).
 * @param data         JSON string representing the business document
 *                     (e.g. a serialised {@code PurchaseOrder}).
 */
public record GenerateRequest(
        @NotNull(message = "EDI format must be specified")
        EdiFormat format,

        @NotBlank(message = "Document type must not be blank")
        String documentType,

        @NotBlank(message = "Sender ID must not be blank")
        String senderId,

        @NotBlank(message = "Receiver ID must not be blank")
        String receiverId,

        @NotBlank(message = "Document data must not be blank")
        String data
) {}
