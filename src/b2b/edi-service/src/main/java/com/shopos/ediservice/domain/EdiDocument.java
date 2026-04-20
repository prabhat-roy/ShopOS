package com.shopos.ediservice.domain;

import java.time.Instant;
import java.util.List;

/**
 * Parsed representation of a complete EDI interchange.
 *
 * @param format          The EDI format standard (X12 or EDIFACT).
 * @param documentType    Business document type: PO (850), INVOICE (810), ASN (856), ACK (997).
 * @param senderId        Interchange sender identifier from the ISA/UNB envelope.
 * @param receiverId      Interchange receiver identifier from the ISA/UNB envelope.
 * @param transactionId   Control number from ISA/ST segment.
 * @param segments        All parsed segments in document order.
 * @param rawContent      Original raw EDI string as received.
 * @param parsedAt        Timestamp when this document was parsed.
 */
public record EdiDocument(
        EdiFormat format,
        String documentType,
        String senderId,
        String receiverId,
        String transactionId,
        List<EdiSegment> segments,
        String rawContent,
        Instant parsedAt
) {

    /**
     * Finds the first segment matching the given segment ID.
     *
     * @param segmentId e.g. "ISA", "BEG", "N1"
     * @return the first matching segment, or {@code null} if not found
     */
    public EdiSegment findSegment(String segmentId) {
        return segments.stream()
                .filter(s -> s.id().equals(segmentId))
                .findFirst()
                .orElse(null);
    }

    /**
     * Returns all segments that match the given segment ID.
     */
    public List<EdiSegment> findAllSegments(String segmentId) {
        return segments.stream()
                .filter(s -> s.id().equals(segmentId))
                .toList();
    }
}
