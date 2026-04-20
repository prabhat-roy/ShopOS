package com.shopos.ediservice.domain;

import java.util.List;

/**
 * A single EDI segment — the fundamental building block of an EDI message.
 *
 * @param id       Segment identifier (e.g. ISA, GS, ST, BEG, N1, PO1, CTT, SE, GE, IEA).
 * @param elements Ordered list of element values within the segment (index 0 is element 1).
 */
public record EdiSegment(String id, List<String> elements) {

    /**
     * Returns element at position {@code index} (1-based as per EDI convention),
     * or an empty string when the index is out of range.
     */
    public String element(int index) {
        int zeroBase = index - 1;
        if (zeroBase < 0 || zeroBase >= elements.size()) {
            return "";
        }
        return elements.get(zeroBase);
    }
}
