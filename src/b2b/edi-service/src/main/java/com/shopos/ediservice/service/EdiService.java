package com.shopos.ediservice.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.shopos.ediservice.domain.*;
import com.shopos.ediservice.dto.EdiResponse;
import com.shopos.ediservice.dto.GenerateRequest;
import com.shopos.ediservice.dto.ParseRequest;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.util.ArrayList;
import java.util.List;

/**
 * Orchestrates EDI parsing, generation and validation for all supported formats.
 */
@Slf4j
@Service
public class EdiService {

    private final X12Parser x12Parser;
    private final X12Generator x12Generator;
    private final ObjectMapper objectMapper;

    public EdiService(X12Parser x12Parser, X12Generator x12Generator) {
        this.x12Parser = x12Parser;
        this.x12Generator = x12Generator;
        this.objectMapper = new ObjectMapper().registerModule(new JavaTimeModule());
    }

    // -------------------------------------------------------------------------
    // Parse
    // -------------------------------------------------------------------------

    /**
     * Parses a raw EDI message and returns a structured response.
     */
    public EdiResponse parse(ParseRequest request) {
        List<String> validationErrors = validateFormat(request.content(), request.format());
        if (!validationErrors.isEmpty()) {
            return EdiResponse.failure(null, request.format(), validationErrors);
        }

        try {
            EdiDocument doc = switch (request.format()) {
                case X12 -> x12Parser.parseX12(request.content());
                case EDIFACT -> throw new UnsupportedOperationException(
                        "EDIFACT parsing is not yet implemented in this release");
            };

            Object parsedDocument = null;
            if ("PO".equals(doc.documentType())) {
                parsedDocument = extractPurchaseOrder(doc);
            }

            return EdiResponse.success(
                    doc.documentType(),
                    doc.format(),
                    doc.transactionId(),
                    parsedDocument,
                    doc.segments().size()
            );

        } catch (IllegalArgumentException ex) {
            log.warn("EDI parse failed: {}", ex.getMessage());
            return EdiResponse.failure(null, request.format(), List.of(ex.getMessage()));
        } catch (UnsupportedOperationException ex) {
            return EdiResponse.failure(null, request.format(), List.of(ex.getMessage()));
        } catch (Exception ex) {
            log.error("Unexpected error during EDI parse", ex);
            return EdiResponse.failure(null, request.format(),
                    List.of("Internal parse error: " + ex.getMessage()));
        }
    }

    // -------------------------------------------------------------------------
    // Generate
    // -------------------------------------------------------------------------

    /**
     * Generates a raw EDI string from a business document.
     *
     * @return raw EDI content
     * @throws IllegalArgumentException for unsupported format/documentType combinations
     */
    public String generate(GenerateRequest request) {
        if (request.format() == EdiFormat.EDIFACT) {
            throw new UnsupportedOperationException(
                    "EDIFACT generation is not yet implemented in this release");
        }

        if (!"PO".equalsIgnoreCase(request.documentType())) {
            throw new UnsupportedOperationException(
                    "Generation is currently supported only for document type PO (850)");
        }

        try {
            PurchaseOrder po = objectMapper.readValue(request.data(), PurchaseOrder.class);
            return x12Generator.generatePO(po, request.senderId(), request.receiverId());
        } catch (Exception ex) {
            log.error("EDI generation failed", ex);
            throw new IllegalArgumentException("Failed to generate EDI: " + ex.getMessage(), ex);
        }
    }

    // -------------------------------------------------------------------------
    // Validate
    // -------------------------------------------------------------------------

    /**
     * Validates an EDI message for structural correctness.
     *
     * @return list of validation error messages; empty when the document is valid
     */
    public List<String> validateFormat(String content, EdiFormat format) {
        List<String> errors = new ArrayList<>();

        if (content == null || content.isBlank()) {
            errors.add("EDI content must not be blank");
            return errors;
        }

        if (format == EdiFormat.X12) {
            validateX12Structure(content.strip(), errors);
        } else if (format == EdiFormat.EDIFACT) {
            validateEdifactStructure(content.strip(), errors);
        }

        return errors;
    }

    // -------------------------------------------------------------------------
    // Extract business document from parsed EdiDocument
    // -------------------------------------------------------------------------

    /**
     * Extracts a {@link PurchaseOrder} from a parsed X12 850 {@link EdiDocument}.
     *
     * @throws IllegalArgumentException when the document is not a PO or is missing required segments
     */
    public PurchaseOrder extractPurchaseOrder(EdiDocument doc) {
        if (doc == null) {
            throw new IllegalArgumentException("EdiDocument must not be null");
        }
        if (!"PO".equals(doc.documentType())) {
            throw new IllegalArgumentException(
                    "Cannot extract PurchaseOrder from document type: " + doc.documentType());
        }

        // BEG — purchase order number and date
        EdiSegment beg = doc.findSegment("BEG");
        String poNumber = beg != null ? beg.element(3) : "UNKNOWN";
        LocalDate orderDate = parseOrderDate(beg != null ? beg.element(5) : null);

        // N1 — party names
        String buyer = "";
        String vendor = "";
        for (EdiSegment n1 : doc.findAllSegments("N1")) {
            String qualifier = n1.element(1);
            String name = n1.element(2);
            if ("BY".equalsIgnoreCase(qualifier)) {
                buyer = name;
            } else if ("SE".equalsIgnoreCase(qualifier) || "VN".equalsIgnoreCase(qualifier)) {
                vendor = name;
            }
        }

        // PO1 — line items
        List<EdiSegment> po1Segments = doc.findAllSegments("PO1");
        List<OrderLine> items = new ArrayList<>();
        for (EdiSegment po1 : po1Segments) {
            int lineNumber = parseIntSafe(po1.element(1));
            BigDecimal quantity = parseBigDecimalSafe(po1.element(2));
            String uom = po1.element(3).isBlank() ? "EA" : po1.element(3);
            BigDecimal unitPrice = parseBigDecimalSafe(po1.element(4));

            // Product IDs: elements 6/7 = qualifier/id pairs
            String productId = "";
            String sku = "";
            List<String> elems = po1.elements();
            for (int i = 5; i < elems.size() - 1; i += 2) {
                String qualifier = elems.get(i);
                String value = elems.get(i + 1);
                if ("BP".equalsIgnoreCase(qualifier)) {
                    productId = value;
                } else if ("VP".equalsIgnoreCase(qualifier) || "SK".equalsIgnoreCase(qualifier)) {
                    sku = value;
                }
            }

            items.add(new OrderLine(lineNumber, productId, sku, quantity, unitPrice, uom));
        }

        // Currency from CUR segment if present
        EdiSegment cur = doc.findSegment("CUR");
        String currency = (cur != null && !cur.element(2).isBlank()) ? cur.element(2) : "USD";

        return PurchaseOrder.of(poNumber, buyer, vendor, orderDate, items, currency);
    }

    // -------------------------------------------------------------------------
    // private helpers
    // -------------------------------------------------------------------------

    private void validateX12Structure(String content, List<String> errors) {
        if (!content.startsWith("ISA")) {
            errors.add("X12 interchange must begin with ISA segment");
        }
        if (!content.contains("~")) {
            errors.add("X12 interchange must use '~' as segment terminator");
        }
        if (!content.contains("GS")) {
            errors.add("Missing GS (Functional Group Header) segment");
        }
        if (!content.contains("GE")) {
            errors.add("Missing GE (Functional Group Trailer) segment");
        }
        if (!content.contains("IEA")) {
            errors.add("Missing IEA (Interchange Control Trailer) segment");
        }
        if (!content.contains("ST")) {
            errors.add("Missing ST (Transaction Set Header) segment");
        }
        if (!content.contains("SE")) {
            errors.add("Missing SE (Transaction Set Trailer) segment");
        }

        // ISA must be at least 106 chars
        if (content.length() < 106) {
            errors.add("ISA segment appears to be truncated (must be at least 106 characters)");
        }
    }

    private void validateEdifactStructure(String content, List<String> errors) {
        if (!content.startsWith("UNB") && !content.startsWith("UNA")) {
            errors.add("EDIFACT interchange must begin with UNA or UNB segment");
        }
        if (!content.contains("UNZ")) {
            errors.add("Missing UNZ (Interchange Trailer) segment");
        }
        if (!content.contains("UNH")) {
            errors.add("Missing UNH (Message Header) segment");
        }
        if (!content.contains("UNT")) {
            errors.add("Missing UNT (Message Trailer) segment");
        }
    }

    private LocalDate parseOrderDate(String dateStr) {
        if (dateStr == null || dateStr.isBlank()) {
            return LocalDate.now();
        }
        try {
            if (dateStr.length() == 8) {
                return LocalDate.parse(dateStr, DateTimeFormatter.ofPattern("yyyyMMdd"));
            }
            if (dateStr.length() == 6) {
                return LocalDate.parse(dateStr, DateTimeFormatter.ofPattern("yyMMdd"));
            }
        } catch (DateTimeParseException ex) {
            log.warn("Could not parse order date '{}', defaulting to today", dateStr);
        }
        return LocalDate.now();
    }

    private int parseIntSafe(String value) {
        try {
            return Integer.parseInt(value.trim());
        } catch (NumberFormatException ex) {
            return 0;
        }
    }

    private BigDecimal parseBigDecimalSafe(String value) {
        try {
            return new BigDecimal(value.trim());
        } catch (NumberFormatException ex) {
            return BigDecimal.ZERO;
        }
    }
}
