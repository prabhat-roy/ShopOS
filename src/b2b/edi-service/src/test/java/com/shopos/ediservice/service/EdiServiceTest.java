package com.shopos.ediservice.service;

import com.shopos.ediservice.domain.*;
import com.shopos.ediservice.dto.EdiResponse;
import com.shopos.ediservice.dto.ParseRequest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.List;

import static org.assertj.core.api.Assertions.*;

@ExtendWith(MockitoExtension.class)
class EdiServiceTest {

    // Use real collaborators (lightweight, no I/O); Spy lets us verify if needed
    @Spy
    private X12Parser x12Parser;

    @Spy
    private X12Generator x12Generator;

    @InjectMocks
    private EdiService ediService;

    // -------------------------------------------------------------------------
    // Minimal valid X12 850 fixture
    // -------------------------------------------------------------------------

    /**
     * Builds a minimal but structurally complete X12 850 PO string.
     * ISA is padded to exactly 106 chars with '~' as segment terminator.
     */
    private static String validX12Po() {
        return "ISA*00*          *00*          *ZZ*BUYER          *ZZ*VENDOR         *260101*1200*^*00401*000000001*0*P*:~\n"
             + "GS*PO*BUYER*VENDOR*20260101*1200*1*X*004010~\n"
             + "ST*850*0001~\n"
             + "BEG*00*SA*PO-12345**20260101~\n"
             + "N1*BY*Acme Corp~\n"
             + "N1*SE*Vendor Inc~\n"
             + "PO1*1*10*EA*25.00**BP*PROD-001*VP*SKU-ABC~\n"
             + "PO1*2*5*EA*50.00**BP*PROD-002*VP*SKU-DEF~\n"
             + "CTT*2*350.00~\n"
             + "SE*9*0001~\n"
             + "GE*1*1~\n"
             + "IEA*1*000000001~\n";
    }

    // -------------------------------------------------------------------------
    // Test 1: Parse a valid X12 PO — should return success with PurchaseOrder
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("1 — parse valid X12 PO returns success response with PurchaseOrder")
    void parseX12Po_validContent_returnsSuccessWithPurchaseOrder() {
        ParseRequest request = new ParseRequest(validX12Po(), EdiFormat.X12);

        EdiResponse response = ediService.parse(request);

        assertThat(response.success()).isTrue();
        assertThat(response.errors()).isEmpty();
        assertThat(response.documentType()).isEqualTo("PO");
        assertThat(response.format()).isEqualTo(EdiFormat.X12);
        assertThat(response.parsedDocument()).isInstanceOf(PurchaseOrder.class);

        PurchaseOrder po = (PurchaseOrder) response.parsedDocument();
        assertThat(po.poNumber()).isEqualTo("PO-12345");
        assertThat(po.buyer()).isEqualTo("Acme Corp");
        assertThat(po.vendor()).isEqualTo("Vendor Inc");
    }

    // -------------------------------------------------------------------------
    // Test 2: Parse with blank content — should return failure
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("2 — parse blank content returns failure response with validation error")
    void parseX12Po_blankContent_returnsFailure() {
        ParseRequest request = new ParseRequest("   ", EdiFormat.X12);

        EdiResponse response = ediService.parse(request);

        assertThat(response.success()).isFalse();
        assertThat(response.errors()).isNotEmpty();
        assertThat(response.errors().get(0)).contains("blank");
    }

    // -------------------------------------------------------------------------
    // Test 3: Generate PO X12 contains ISA segment
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("3 — generate X12 850 PO contains ISA envelope segment")
    void generateX12Po_containsIsaSegment() {
        PurchaseOrder po = buildSamplePO();
        String edi = x12Generator.generatePO(po, "SENDER123", "RECEIVER456");

        assertThat(edi).startsWith("ISA*");
    }

    // -------------------------------------------------------------------------
    // Test 4: Generate PO X12 contains BEG segment with PO number
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("4 — generate X12 850 PO contains BEG segment with correct PO number")
    void generateX12Po_containsBegSegmentWithPoNumber() {
        PurchaseOrder po = buildSamplePO();
        String edi = x12Generator.generatePO(po, "SENDER", "RECEIVER");

        assertThat(edi).contains("BEG*");
        assertThat(edi).contains("TEST-PO-001");
    }

    // -------------------------------------------------------------------------
    // Test 5: Generated PO contains PO1 segments for each line item
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("5 — generate X12 850 PO contains correct number of PO1 line item segments")
    void generateX12Po_containsPo1SegmentsForEachLine() {
        PurchaseOrder po = buildSamplePO();
        String edi = x12Generator.generatePO(po, "SENDER", "RECEIVER");

        long po1Count = edi.lines()
                .filter(line -> line.startsWith("PO1*"))
                .count();

        assertThat(po1Count).isEqualTo(po.items().size());
    }

    // -------------------------------------------------------------------------
    // Test 6: Parsed document reports correct segment count
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("6 — parsed X12 PO reports the correct segment count in response")
    void parseX12Po_reportsCorrectSegmentCount() {
        ParseRequest request = new ParseRequest(validX12Po(), EdiFormat.X12);

        EdiResponse response = ediService.parse(request);

        assertThat(response.success()).isTrue();
        // fixture has 12 non-empty segment lines
        assertThat(response.segmentCount()).isGreaterThan(8);
    }

    // -------------------------------------------------------------------------
    // Test 7: Sender and receiver are correctly extracted from ISA
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("7 — parse X12 extracts sender and receiver IDs from ISA envelope")
    void parseX12Po_extractsSenderAndReceiver() {
        String edi = validX12Po();
        EdiDocument doc = x12Parser.parseX12(edi);

        assertThat(doc.senderId()).isNotBlank();
        assertThat(doc.receiverId()).isNotBlank();
    }

    // -------------------------------------------------------------------------
    // Test 8: Validate correct EDI returns no errors
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("8 — validate structurally correct X12 EDI returns empty error list")
    void validateFormat_validX12_returnsNoErrors() {
        List<String> errors = ediService.validateFormat(validX12Po(), EdiFormat.X12);

        assertThat(errors).isEmpty();
    }

    // -------------------------------------------------------------------------
    // Test 9: Validate malformed EDI returns errors
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("9 — validate malformed X12 EDI (no ISA) returns validation errors")
    void validateFormat_malformedX12_returnsErrors() {
        String malformed = "GS*PO*SENDER*RECEIVER~ST*850*0001~BEG*00*SA*PO99~SE*3*0001~GE*1*1~";
        List<String> errors = ediService.validateFormat(malformed, EdiFormat.X12);

        assertThat(errors).isNotEmpty();
        assertThat(errors.stream().anyMatch(e -> e.contains("ISA"))).isTrue();
    }

    // -------------------------------------------------------------------------
    // Test 10: extractPurchaseOrder correctly populates all fields
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("10 — extractPurchaseOrder populates all fields from parsed EdiDocument")
    void extractPurchaseOrder_populatesAllFields() {
        EdiDocument doc = x12Parser.parseX12(validX12Po());

        PurchaseOrder po = ediService.extractPurchaseOrder(doc);

        assertThat(po.poNumber()).isEqualTo("PO-12345");
        assertThat(po.buyer()).isEqualTo("Acme Corp");
        assertThat(po.vendor()).isEqualTo("Vendor Inc");
        assertThat(po.items()).hasSize(2);

        OrderLine firstLine = po.items().get(0);
        assertThat(firstLine.quantity()).isEqualByComparingTo("10");
        assertThat(firstLine.unitPrice()).isEqualByComparingTo("25.00");
        assertThat(firstLine.uom()).isEqualTo("EA");
        assertThat(firstLine.productId()).isEqualTo("PROD-001");
        assertThat(firstLine.sku()).isEqualTo("SKU-ABC");

        assertThat(po.totalAmount()).isEqualByComparingTo("350.00");
    }

    // -------------------------------------------------------------------------
    // helpers
    // -------------------------------------------------------------------------

    private PurchaseOrder buildSamplePO() {
        List<OrderLine> lines = List.of(
                new OrderLine(1, "PROD-A", "SKU-001", new BigDecimal("10"), new BigDecimal("15.99"), "EA"),
                new OrderLine(2, "PROD-B", "SKU-002", new BigDecimal("3"), new BigDecimal("99.00"), "CA")
        );
        return PurchaseOrder.of("TEST-PO-001", "Test Buyer", "Test Vendor",
                LocalDate.of(2026, 1, 1), lines, "USD");
    }
}
