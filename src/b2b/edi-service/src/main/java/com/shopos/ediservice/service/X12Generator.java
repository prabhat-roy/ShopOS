package com.shopos.ediservice.service;

import com.shopos.ediservice.domain.OrderLine;
import com.shopos.ediservice.domain.PurchaseOrder;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.time.format.DateTimeFormatter;

/**
 * Generates ANSI X12 EDI 850 Purchase Order messages from business domain objects.
 *
 * <p>Segment terminator: {@code ~}<br>
 * Element separator:  {@code *}<br>
 * Sub-element separator: {@code :}
 */
@Slf4j
@Service
public class X12Generator {

    private static final String ELEM_SEP = "*";
    private static final String SEG_TERM = "~\n";
    private static final DateTimeFormatter DATE_FMT = DateTimeFormatter.ofPattern("yyMMdd");
    private static final DateTimeFormatter TIME_FMT = DateTimeFormatter.ofPattern("HHmm");

    /**
     * Generates a complete X12 850 EDI string for the given purchase order.
     *
     * @param po         the purchase order business object
     * @param senderId   ISA06 sender qualifier + ID (padded to 15 chars)
     * @param receiverId ISA08 receiver qualifier + ID (padded to 15 chars)
     * @return raw X12 string ready for transmission
     */
    public String generatePO(PurchaseOrder po, String senderId, String receiverId) {
        String today = LocalDate.now().format(DATE_FMT);
        String now = java.time.LocalTime.now().format(TIME_FMT);
        String icn = generateControlNumber();  // interchange control number

        StringBuilder sb = new StringBuilder();

        // ISA — Interchange Control Header (exactly 106 chars + terminator)
        sb.append(buildIsa(senderId, receiverId, today, now, icn));

        // GS — Functional Group Header
        sb.append("GS").append(ELEM_SEP)
                .append("PO").append(ELEM_SEP)                    // GS01 functional ID
                .append(pad(senderId, 15)).append(ELEM_SEP)       // GS02 sender
                .append(pad(receiverId, 15)).append(ELEM_SEP)     // GS03 receiver
                .append(today).append(ELEM_SEP)                   // GS04 date
                .append(now).append(ELEM_SEP)                     // GS05 time
                .append("1").append(ELEM_SEP)                     // GS06 group control number
                .append("X").append(ELEM_SEP)                     // GS07 responsible agency
                .append("004010")                                  // GS08 version/release
                .append(SEG_TERM);

        // ST — Transaction Set Header
        sb.append("ST").append(ELEM_SEP)
                .append("850").append(ELEM_SEP)   // ST01 transaction set ID
                .append("0001")                   // ST02 control number
                .append(SEG_TERM);

        // BEG — Beginning Segment for Purchase Order
        String orderDateStr = po.orderDate() != null
                ? po.orderDate().format(DateTimeFormatter.ofPattern("yyyyMMdd"))
                : today;
        sb.append("BEG").append(ELEM_SEP)
                .append("00").append(ELEM_SEP)             // BEG01 transaction set purpose: original
                .append("SA").append(ELEM_SEP)             // BEG02 PO type: standard
                .append(po.poNumber()).append(ELEM_SEP)    // BEG03 PO number
                .append("").append(ELEM_SEP)               // BEG04 release number (not used)
                .append(orderDateStr)                      // BEG05 date
                .append(SEG_TERM);

        // CUR — Currency (if not USD)
        if (po.currency() != null && !po.currency().equalsIgnoreCase("USD")) {
            sb.append("CUR").append(ELEM_SEP)
                    .append("BY").append(ELEM_SEP)         // CUR01 entity identifier
                    .append(po.currency().toUpperCase())   // CUR02 currency code
                    .append(SEG_TERM);
        }

        // N1*BY — Buyer party identification
        if (po.buyer() != null && !po.buyer().isBlank()) {
            sb.append("N1").append(ELEM_SEP)
                    .append("BY").append(ELEM_SEP)         // N101 entity identifier: buyer
                    .append(po.buyer())                    // N102 name
                    .append(SEG_TERM);
        }

        // N1*SE — Seller/vendor party identification
        if (po.vendor() != null && !po.vendor().isBlank()) {
            sb.append("N1").append(ELEM_SEP)
                    .append("SE").append(ELEM_SEP)         // N101 entity identifier: seller
                    .append(po.vendor())                   // N102 name
                    .append(SEG_TERM);
        }

        // PO1 — Line Item Detail loops
        int lineCount = 0;
        for (OrderLine line : po.items()) {
            lineCount++;
            sb.append("PO1").append(ELEM_SEP)
                    .append(lineCount).append(ELEM_SEP)            // PO101 line number
                    .append(line.quantity().toPlainString()).append(ELEM_SEP)  // PO102 quantity
                    .append(line.uom() != null ? line.uom() : "EA").append(ELEM_SEP) // PO103 UOM
                    .append(line.unitPrice().toPlainString()).append(ELEM_SEP) // PO104 unit price
                    .append("").append(ELEM_SEP)                   // PO105 (reserved)
                    .append("BP").append(ELEM_SEP)                 // PO106 product ID qualifier: buyer part
                    .append(line.productId() != null ? line.productId() : "").append(ELEM_SEP) // PO107
                    .append("VP").append(ELEM_SEP)                 // PO108 product ID qualifier: vendor part
                    .append(line.sku() != null ? line.sku() : "")  // PO109 vendor part number
                    .append(SEG_TERM);
        }

        // CTT — Transaction Totals
        sb.append("CTT").append(ELEM_SEP)
                .append(lineCount).append(ELEM_SEP)             // CTT01 number of line items
                .append(po.totalAmount().toPlainString())       // CTT02 hash total (sum of quantities * prices)
                .append(SEG_TERM);

        // SE — Transaction Set Trailer
        // SE01 = segment count (all segments from ST to SE inclusive)
        // We count segments: ST + BEG + (CUR?) + N1s + PO1s + CTT + SE
        int segmentCount = countSegments(sb.toString()) + 1; // +1 for SE itself
        sb.append("SE").append(ELEM_SEP)
                .append(segmentCount).append(ELEM_SEP)   // SE01 segment count
                .append("0001")                          // SE02 transaction set control number
                .append(SEG_TERM);

        // GE — Functional Group Trailer
        sb.append("GE").append(ELEM_SEP)
                .append("1").append(ELEM_SEP)   // GE01 number of transaction sets
                .append("1")                    // GE02 group control number
                .append(SEG_TERM);

        // IEA — Interchange Control Trailer
        sb.append("IEA").append(ELEM_SEP)
                .append("1").append(ELEM_SEP)   // IEA01 number of functional groups
                .append(icn)                    // IEA02 interchange control number
                .append(SEG_TERM);

        log.debug("Generated X12 850 for PO {} — {} line items", po.poNumber(), lineCount);
        return sb.toString();
    }

    // -------------------------------------------------------------------------
    // helpers
    // -------------------------------------------------------------------------

    private String buildIsa(String senderId, String receiverId,
                            String date, String time, String icn) {
        // ISA must be exactly 106 characters (not counting the terminator)
        return "ISA" + ELEM_SEP
                + "00" + ELEM_SEP               // ISA01 auth info qualifier
                + pad("", 10) + ELEM_SEP        // ISA02 auth info
                + "00" + ELEM_SEP               // ISA03 security info qualifier
                + pad("", 10) + ELEM_SEP        // ISA04 security info
                + "ZZ" + ELEM_SEP               // ISA05 sender ID qualifier
                + pad(senderId, 15) + ELEM_SEP  // ISA06 sender ID
                + "ZZ" + ELEM_SEP               // ISA07 receiver ID qualifier
                + pad(receiverId, 15) + ELEM_SEP// ISA08 receiver ID
                + date + ELEM_SEP               // ISA09 interchange date
                + time + ELEM_SEP               // ISA10 interchange time
                + "^" + ELEM_SEP                // ISA11 repetition separator
                + "00401" + ELEM_SEP            // ISA12 control version
                + pad(icn, 9) + ELEM_SEP        // ISA13 interchange control number
                + "0" + ELEM_SEP                // ISA14 acknowledgement requested
                + "P"                           // ISA15 usage indicator: Production
                + ":" + SEG_TERM;               // ISA16 sub-element separator + segment terminator
    }

    private String pad(String value, int length) {
        if (value == null) {
            value = "";
        }
        if (value.length() >= length) {
            return value.substring(0, length);
        }
        return value + " ".repeat(length - value.length());
    }

    private String generateControlNumber() {
        // Use current time in millis modulo 999999999 for a 9-digit control number
        long num = System.currentTimeMillis() % 999_999_999L;
        return String.format("%09d", num);
    }

    /**
     * Counts segments delimited by {@code ~} in the generated content so far
     * (to compute the SE01 segment count). Includes all segments from ST onward.
     */
    private int countSegments(String content) {
        int stIndex = content.indexOf("ST*");
        if (stIndex < 0) {
            return 0;
        }
        String fromSt = content.substring(stIndex);
        int count = 0;
        for (char c : fromSt.toCharArray()) {
            if (c == '~') {
                count++;
            }
        }
        return count;
    }
}
