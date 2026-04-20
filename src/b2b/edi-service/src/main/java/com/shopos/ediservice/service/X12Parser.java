package com.shopos.ediservice.service;

import com.shopos.ediservice.domain.EdiDocument;
import com.shopos.ediservice.domain.EdiFormat;
import com.shopos.ediservice.domain.EdiSegment;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * Parses ANSI X12 EDI messages into {@link EdiDocument} objects.
 *
 * <p>X12 uses:
 * <ul>
 *   <li>{@code ~} as segment terminator</li>
 *   <li>{@code *} as element separator</li>
 *   <li>{@code :} as sub-element separator</li>
 * </ul>
 *
 * <p>The ISA envelope is always 106 characters; terminators are read dynamically
 * from ISA16 (element separator) and the character immediately after ISA (segment terminator).
 */
@Slf4j
@Service
public class X12Parser {

    private static final char DEFAULT_ELEMENT_SEPARATOR = '*';
    private static final char DEFAULT_SEGMENT_TERMINATOR = '~';

    /**
     * Parses a raw X12 EDI string into an {@link EdiDocument}.
     *
     * @param content raw X12 EDI content
     * @return fully parsed EdiDocument
     * @throws IllegalArgumentException when content is missing the ISA envelope
     */
    public EdiDocument parseX12(String content) {
        if (content == null || content.isBlank()) {
            throw new IllegalArgumentException("X12 content must not be blank");
        }

        String normalized = content.strip();

        // The ISA segment is always 106 chars; detect separators from it
        if (!normalized.startsWith("ISA")) {
            throw new IllegalArgumentException("X12 content must begin with an ISA segment");
        }

        char elementSep = normalized.charAt(3);          // position after "ISA"
        char segmentTerm = normalized.charAt(105);       // last char of ISA is the segment terminator

        log.debug("X12 parsing — element separator: '{}', segment terminator: '{}'",
                elementSep, segmentTerm);

        // Split into raw segment strings
        String[] rawSegments = normalized.split("\\" + segmentTerm);

        List<EdiSegment> segments = new ArrayList<>();
        String senderId = "";
        String receiverId = "";
        String transactionId = "";
        String documentType = "UNKNOWN";

        for (String rawSeg : rawSegments) {
            String trimmed = rawSeg.strip();
            if (trimmed.isEmpty()) {
                continue;
            }

            String[] parts = trimmed.split("\\" + elementSep, -1);
            String segId = parts[0].trim();
            List<String> elements = new ArrayList<>();
            for (int i = 1; i < parts.length; i++) {
                elements.add(parts[i].trim());
            }

            EdiSegment segment = new EdiSegment(segId, elements);
            segments.add(segment);

            switch (segId) {
                case "ISA" -> {
                    // ISA06 = sender, ISA08 = receiver, ISA13 = interchange control number
                    senderId = elements.size() > 5 ? elements.get(5).trim() : "";
                    receiverId = elements.size() > 7 ? elements.get(7).trim() : "";
                    if (transactionId.isEmpty()) {
                        transactionId = elements.size() > 12 ? elements.get(12).trim() : "";
                    }
                }
                case "ST" -> {
                    // ST01 = transaction set ID (850=PO, 810=INVOICE, 856=ASN, 997=ACK)
                    String txSetId = elements.isEmpty() ? "" : elements.get(0).trim();
                    documentType = resolveDocumentType(txSetId);
                    // ST02 = transaction set control number
                    if (elements.size() > 1) {
                        transactionId = elements.get(1).trim();
                    }
                }
                default -> {
                    // nothing extra — segment already collected
                }
            }
        }

        log.debug("X12 parse complete — {} segments, documentType={}, sender={}, receiver={}",
                segments.size(), documentType, senderId, receiverId);

        return new EdiDocument(
                EdiFormat.X12,
                documentType,
                senderId,
                receiverId,
                transactionId,
                segments,
                content,
                Instant.now()
        );
    }

    // -------------------------------------------------------------------------
    // helpers
    // -------------------------------------------------------------------------

    /**
     * Maps X12 transaction set IDs to human-readable document type names.
     */
    private String resolveDocumentType(String txSetId) {
        return switch (txSetId) {
            case "850" -> "PO";
            case "810" -> "INVOICE";
            case "856" -> "ASN";
            case "997" -> "ACK";
            case "855" -> "PO_ACK";
            case "860" -> "PO_CHANGE";
            default -> txSetId.isEmpty() ? "UNKNOWN" : txSetId;
        };
    }
}
