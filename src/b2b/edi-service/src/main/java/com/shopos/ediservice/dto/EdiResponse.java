package com.shopos.ediservice.dto;

import com.shopos.ediservice.domain.EdiFormat;

import java.util.List;

/**
 * Response returned by the EDI parse and validate endpoints.
 *
 * @param documentType    Detected or requested business document type.
 * @param format          EDI format standard.
 * @param transactionId   Control/transaction number extracted from the interchange.
 * @param parsedDocument  Structured business object (e.g. {@code PurchaseOrder}); may be null on error.
 * @param segmentCount    Total number of segments in the parsed document.
 * @param success         {@code true} when parsing completed without fatal errors.
 * @param errors          List of validation or parse error messages; empty on success.
 */
public record EdiResponse(
        String documentType,
        EdiFormat format,
        String transactionId,
        Object parsedDocument,
        int segmentCount,
        boolean success,
        List<String> errors
) {

    public static EdiResponse success(
            String documentType,
            EdiFormat format,
            String transactionId,
            Object parsedDocument,
            int segmentCount) {

        return new EdiResponse(documentType, format, transactionId, parsedDocument,
                segmentCount, true, List.of());
    }

    public static EdiResponse failure(
            String documentType,
            EdiFormat format,
            List<String> errors) {

        return new EdiResponse(documentType, format, null, null, 0, false, errors);
    }
}
