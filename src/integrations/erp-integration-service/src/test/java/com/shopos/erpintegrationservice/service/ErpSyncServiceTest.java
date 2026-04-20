package com.shopos.erpintegrationservice.service;

import com.shopos.erpintegrationservice.domain.ErpLineItem;
import com.shopos.erpintegrationservice.domain.ErpOrder;
import com.shopos.erpintegrationservice.domain.ErpSystem;
import com.shopos.erpintegrationservice.domain.SyncStatus;
import com.shopos.erpintegrationservice.dto.SyncInventoryRequest;
import com.shopos.erpintegrationservice.dto.SyncOrderRequest;
import com.shopos.erpintegrationservice.dto.SyncResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@DisplayName("ErpSyncService unit tests")
class ErpSyncServiceTest {

    @Mock
    private ErpTranslator erpTranslator;

    @InjectMocks
    private ErpSyncService erpSyncService;

    private ErpOrder sampleErpOrder;

    @BeforeEach
    void setUp() {
        ErpLineItem lineItem = new ErpLineItem(
                "MATNR-prod-001", "prod-001", "SKU-001", 2,
                new BigDecimal("149.99"), "EA");

        sampleErpOrder = new ErpOrder(
                "VBELN-order-001", "order-001", ErpSystem.SAP,
                "KUNNR-customer-001", List.of(lineItem),
                new BigDecimal("299.98"), "USD",
                SyncStatus.SUCCESS, Instant.now());
    }

    // -------------------------------------------------------------------------
    // syncOrder tests
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("syncOrder SAP — returns SUCCESS with one record processed")
    void syncOrder_sap_success() {
        when(erpTranslator.translateOrderToErp(any(), eq(ErpSystem.SAP))).thenReturn(sampleErpOrder);

        SyncOrderRequest request = new SyncOrderRequest("order-001", ErpSystem.SAP, Map.of("shopOsCustomerId", "C10042"));
        SyncResponse response = erpSyncService.syncOrder(request);

        assertThat(response.syncId()).isNotNull();
        assertThat(response.status()).isEqualTo(SyncStatus.SUCCESS);
        assertThat(response.recordsProcessed()).isEqualTo(1);
        assertThat(response.errors()).isEmpty();
        assertThat(response.completedAt()).isNotNull();
    }

    @Test
    @DisplayName("syncOrder Oracle — returns SUCCESS and assigns a unique sync ID")
    void syncOrder_oracle_success() {
        ErpOrder oracleOrder = new ErpOrder(
                "ORDER_NUMBER-order-002", "order-002", ErpSystem.ORACLE,
                "CUSTOMER_ID-cust-02", List.of(), new BigDecimal("99.99"), "USD",
                SyncStatus.SUCCESS, Instant.now());

        when(erpTranslator.translateOrderToErp(any(), eq(ErpSystem.ORACLE))).thenReturn(oracleOrder);

        SyncOrderRequest request = new SyncOrderRequest("order-002", ErpSystem.ORACLE, null);
        SyncResponse first = erpSyncService.syncOrder(request);
        SyncResponse second = erpSyncService.syncOrder(request);

        assertThat(first.status()).isEqualTo(SyncStatus.SUCCESS);
        assertThat(second.status()).isEqualTo(SyncStatus.SUCCESS);
        assertThat(first.syncId()).isNotEqualTo(second.syncId());
    }

    // -------------------------------------------------------------------------
    // syncInventory tests
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("syncInventory — processes each product and returns SUCCESS")
    void syncInventory_success() {
        when(erpTranslator.translateInventory(any(), eq(ErpSystem.SAP)))
                .thenAnswer(invocation -> new com.shopos.erpintegrationservice.domain.ErpInventory(
                        "MATNR-prod-001", "prod-001", "SKU-001", 50, "WH-01", Instant.now()));

        SyncInventoryRequest request = new SyncInventoryRequest(
                "WH-01", ErpSystem.SAP, List.of("prod-001", "prod-002"));
        SyncResponse response = erpSyncService.syncInventory(request);

        assertThat(response.status()).isEqualTo(SyncStatus.SUCCESS);
        assertThat(response.recordsProcessed()).isEqualTo(2);
        assertThat(response.errors()).isEmpty();
    }

    // -------------------------------------------------------------------------
    // getSyncStatus tests
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("getSyncStatus — returns status for a known sync ID")
    void getSyncStatus_found() {
        when(erpTranslator.translateOrderToErp(any(), any())).thenReturn(sampleErpOrder);

        SyncOrderRequest request = new SyncOrderRequest("order-100", ErpSystem.SAP, null);
        SyncResponse created = erpSyncService.syncOrder(request);

        Optional<SyncStatus> status = erpSyncService.getSyncStatus(created.syncId());

        assertThat(status).isPresent();
        assertThat(status.get()).isEqualTo(SyncStatus.SUCCESS);
    }

    @Test
    @DisplayName("getSyncStatus — returns empty for an unknown sync ID")
    void getSyncStatus_notFound() {
        Optional<SyncStatus> status = erpSyncService.getSyncStatus(UUID.randomUUID());
        assertThat(status).isEmpty();
    }

    // -------------------------------------------------------------------------
    // Field mapping delegation tests
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("field mapping SAP — contains VBELN as orderId mapping")
    void fieldMapping_sap_hasVbeln() {
        ErpTranslator realTranslator = new ErpTranslator();
        Map<String, String> mapping = realTranslator.getFieldMapping(ErpSystem.SAP);

        assertThat(mapping).containsEntry("orderId", "VBELN");
        assertThat(mapping).containsEntry("productId", "MATNR");
    }

    @Test
    @DisplayName("field mapping Oracle — contains ORDER_NUMBER as orderId mapping")
    void fieldMapping_oracle_hasOrderNumber() {
        ErpTranslator realTranslator = new ErpTranslator();
        Map<String, String> mapping = realTranslator.getFieldMapping(ErpSystem.ORACLE);

        assertThat(mapping).containsEntry("orderId", "ORDER_NUMBER");
        assertThat(mapping).containsEntry("productId", "ITEM_NUMBER");
    }

    // -------------------------------------------------------------------------
    // ErpTranslator translation tests
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("translateOrderToErp SAP — erpOrderId contains VBELN prefix")
    void translateOrder_sap_erpOrderIdContainsVbeln() {
        ErpTranslator realTranslator = new ErpTranslator();
        Map<String, Object> orderData = Map.of(
                "orderId", "SO-999",
                "totalAmount", "200.00",
                "currency", "USD",
                "customerId", "CUST-1",
                "lineItems", List.of(
                        Map.of("productId", "P1", "sku", "SKU1", "quantity", 1,
                               "unitPrice", "200.00", "uom", "EA")
                )
        );

        ErpOrder result = realTranslator.translateOrderToErp(orderData, ErpSystem.SAP);

        assertThat(result.erpOrderId()).contains("VBELN");
        assertThat(result.shopOsOrderId()).isEqualTo("SO-999");
        assertThat(result.erpSystem()).isEqualTo(ErpSystem.SAP);
        assertThat(result.lineItems()).hasSize(1);
        assertThat(result.lineItems().get(0).erpProductId()).contains("MATNR");
    }

    @Test
    @DisplayName("translateOrderToErp NetSuite — erpOrderId contains tranId prefix")
    void translateOrder_netsuite_erpOrderIdContainsTranId() {
        ErpTranslator realTranslator = new ErpTranslator();
        Map<String, Object> orderData = Map.of(
                "orderId", "NS-500",
                "totalAmount", "300.00",
                "currency", "EUR",
                "lineItems", List.of()
        );

        ErpOrder result = realTranslator.translateOrderToErp(orderData, ErpSystem.NETSUITE);

        assertThat(result.erpOrderId()).contains("tranId");
        assertThat(result.currency()).isEqualTo("EUR");
        assertThat(result.erpSystem()).isEqualTo(ErpSystem.NETSUITE);
    }

    // -------------------------------------------------------------------------
    // getRecentSyncs tests
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("getRecentSyncs — limit caps result size")
    void getRecentSyncs_limitIsRespected() {
        when(erpTranslator.translateOrderToErp(any(), eq(ErpSystem.DYNAMICS)))
                .thenReturn(new ErpOrder("dyn-1", "ord-1", ErpSystem.DYNAMICS,
                        "cust-1", List.of(), BigDecimal.TEN, "USD",
                        SyncStatus.SUCCESS, Instant.now()));

        SyncOrderRequest request = new SyncOrderRequest("ord-x", ErpSystem.DYNAMICS, null);
        for (int i = 0; i < 5; i++) {
            erpSyncService.syncOrder(request);
        }

        List<SyncResponse> results = erpSyncService.getRecentSyncs(ErpSystem.DYNAMICS, 3);

        assertThat(results).hasSize(3);
        assertThat(results).allMatch(r -> r.status() == SyncStatus.SUCCESS);
    }
}
