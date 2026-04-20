package com.shopos.erpintegrationservice.service;

import com.shopos.erpintegrationservice.domain.ErpInventory;
import com.shopos.erpintegrationservice.domain.ErpLineItem;
import com.shopos.erpintegrationservice.domain.ErpOrder;
import com.shopos.erpintegrationservice.domain.ErpSystem;
import com.shopos.erpintegrationservice.domain.SyncStatus;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * Translates domain objects between ShopOS canonical format and ERP-system-specific formats.
 *
 * <p>Each ERP system uses different field names:
 * <ul>
 *   <li>SAP     – VBELN (order doc), MATNR (material number), KUNNR (customer)</li>
 *   <li>Oracle  – ORDER_NUMBER, ITEM_NUMBER, CUSTOMER_ID</li>
 *   <li>NetSuite – tranId, itemId, entity</li>
 *   <li>Dynamics – salesOrderId, productNumber, customerId</li>
 *   <li>Generic – order_id, product_id, customer_id (camelCase fallback)</li>
 * </ul>
 */
@Service
public class ErpTranslator {

    /**
     * Returns the field-name mapping from ShopOS canonical names to ERP-native names.
     *
     * @param erpSystem The target ERP system.
     * @return Immutable map of { shopOsFieldName → erpFieldName }.
     */
    public Map<String, String> getFieldMapping(ErpSystem erpSystem) {
        return switch (erpSystem) {
            case SAP -> Map.of(
                    "orderId",     "VBELN",
                    "productId",   "MATNR",
                    "customerId",  "KUNNR",
                    "quantity",    "KWMENG",
                    "unitPrice",   "NETPR",
                    "currency",    "WAERK",
                    "warehouseId", "LGORT",
                    "sku",         "EAN11",
                    "uom",         "VRKME",
                    "totalAmount", "NETWR"
            );
            case ORACLE -> Map.of(
                    "orderId",     "ORDER_NUMBER",
                    "productId",   "ITEM_NUMBER",
                    "customerId",  "CUSTOMER_ID",
                    "quantity",    "ORDERED_QUANTITY",
                    "unitPrice",   "UNIT_SELLING_PRICE",
                    "currency",    "TRANSACTIONAL_CURR_CODE",
                    "warehouseId", "SHIP_FROM_ORG_ID",
                    "sku",         "ORDERED_ITEM",
                    "uom",         "ORDER_QUANTITY_UOM",
                    "totalAmount", "FLOW_STATUS_CODE"
            );
            case NETSUITE -> Map.of(
                    "orderId",     "tranId",
                    "productId",   "itemId",
                    "customerId",  "entity",
                    "quantity",    "quantity",
                    "unitPrice",   "rate",
                    "currency",    "currency",
                    "warehouseId", "location",
                    "sku",         "itemId",
                    "uom",         "units",
                    "totalAmount", "subtotal"
            );
            case DYNAMICS -> Map.of(
                    "orderId",     "salesOrderId",
                    "productId",   "productNumber",
                    "customerId",  "customerId",
                    "quantity",    "salesQuantity",
                    "unitPrice",   "unitPrice",
                    "currency",    "currencyCode",
                    "warehouseId", "warehouseId",
                    "sku",         "productNumber",
                    "uom",         "unitOfMeasure",
                    "totalAmount", "totalNetAmount"
            );
            case GENERIC -> Map.of(
                    "orderId",     "order_id",
                    "productId",   "product_id",
                    "customerId",  "customer_id",
                    "quantity",    "qty",
                    "unitPrice",   "unit_price",
                    "currency",    "currency",
                    "warehouseId", "warehouse_id",
                    "sku",         "sku",
                    "uom",         "uom",
                    "totalAmount", "total_amount"
            );
        };
    }

    /**
     * Translates a ShopOS order payload (as a raw field map) into an {@link ErpOrder}.
     *
     * <p>The incoming {@code orderData} map is expected to use ShopOS canonical field names.
     * Fields not present fall back to sensible defaults so translation never throws on missing keys.
     *
     * @param orderData Raw ShopOS order fields.
     * @param erpSystem Target ERP system.
     * @return An {@link ErpOrder} populated with ERP-native identifiers.
     */
    public ErpOrder translateOrderToErp(Map<String, Object> orderData, ErpSystem erpSystem) {
        Map<String, String> mapping = getFieldMapping(erpSystem);

        String shopOsOrderId  = getString(orderData, "orderId", "UNKNOWN-" + UUID.randomUUID());
        String erpOrderId     = mapping.getOrDefault("orderId", "order_id") + "-" + shopOsOrderId;
        String customerErpId  = resolveCustomerErpId(orderData, erpSystem, mapping);

        @SuppressWarnings("unchecked")
        List<Map<String, Object>> rawItems =
                (List<Map<String, Object>>) orderData.getOrDefault("lineItems", List.of());

        List<ErpLineItem> lineItems = rawItems.stream()
                .map(item -> translateLineItem(item, erpSystem, mapping))
                .toList();

        BigDecimal totalAmount = parseBigDecimal(orderData, "totalAmount", BigDecimal.ZERO);
        String currency        = getString(orderData, "currency", "USD");

        return new ErpOrder(
                erpOrderId,
                shopOsOrderId,
                erpSystem,
                customerErpId,
                lineItems,
                totalAmount,
                currency,
                SyncStatus.SUCCESS,
                Instant.now()
        );
    }

    /**
     * Translates an {@link ErpOrder} back into a ShopOS-canonical field map.
     *
     * @param erpOrder The ERP order to translate.
     * @return A map using ShopOS canonical field names.
     */
    public Map<String, Object> translateErpToOrder(ErpOrder erpOrder) {
        Map<String, Object> result = new HashMap<>();
        result.put("orderId",       erpOrder.shopOsOrderId());
        result.put("erpOrderId",    erpOrder.erpOrderId());
        result.put("erpSystem",     erpOrder.erpSystem().name());
        result.put("customerId",    erpOrder.customerErpId());
        result.put("totalAmount",   erpOrder.totalAmount());
        result.put("currency",      erpOrder.currency());
        result.put("status",        erpOrder.status().name());
        result.put("syncedAt",      erpOrder.syncedAt().toString());

        List<Map<String, Object>> items = new ArrayList<>();
        for (ErpLineItem li : erpOrder.lineItems()) {
            Map<String, Object> item = new HashMap<>();
            item.put("productId",      li.shopOsProductId());
            item.put("erpProductId",   li.erpProductId());
            item.put("sku",            li.sku());
            item.put("quantity",       li.quantity());
            item.put("unitPrice",      li.unitPrice());
            item.put("uom",            li.uom());
            items.add(item);
        }
        result.put("lineItems", items);
        return result;
    }

    /**
     * Translates a raw ShopOS inventory payload into an {@link ErpInventory}.
     *
     * @param inventoryData Raw inventory fields (shopOsProductId, sku, quantity, warehouseId, lastUpdated).
     * @param erpSystem     Target ERP system.
     * @return An {@link ErpInventory} with ERP-native product identifier.
     */
    public ErpInventory translateInventory(Map<String, Object> inventoryData, ErpSystem erpSystem) {
        Map<String, String> mapping = getFieldMapping(erpSystem);

        String shopOsProductId = getString(inventoryData, "shopOsProductId", "UNKNOWN");
        String erpProductId    = mapping.getOrDefault("productId", "product_id") + "-" + shopOsProductId;
        String sku             = getString(inventoryData, "sku", "");
        int quantity           = parseInteger(inventoryData, "quantity", 0);
        String warehouseId     = getString(inventoryData, "warehouseId", "DEFAULT");
        Instant lastUpdated    = parseInstant(inventoryData, "lastUpdated");

        return new ErpInventory(erpProductId, shopOsProductId, sku, quantity, warehouseId, lastUpdated);
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    private ErpLineItem translateLineItem(Map<String, Object> item,
                                          ErpSystem erpSystem,
                                          Map<String, String> mapping) {
        String shopOsProductId = getString(item, "productId", "UNKNOWN");
        String erpProductId    = mapping.getOrDefault("productId", "product_id") + "-" + shopOsProductId;
        String sku             = getString(item, "sku", "");
        int quantity           = parseInteger(item, "quantity", 1);
        BigDecimal unitPrice   = parseBigDecimal(item, "unitPrice", BigDecimal.ZERO);
        String uom             = getString(item, "uom", "EA");
        return new ErpLineItem(erpProductId, shopOsProductId, sku, quantity, unitPrice, uom);
    }

    private String resolveCustomerErpId(Map<String, Object> orderData,
                                         ErpSystem erpSystem,
                                         Map<String, String> mapping) {
        String shopOsCustomerId = getString(orderData, "customerId", "GUEST");
        return mapping.getOrDefault("customerId", "customer_id") + "-" + shopOsCustomerId;
    }

    private String getString(Map<String, Object> map, String key, String defaultValue) {
        Object val = map.get(key);
        return val != null ? val.toString() : defaultValue;
    }

    private BigDecimal parseBigDecimal(Map<String, Object> map, String key, BigDecimal defaultValue) {
        Object val = map.get(key);
        if (val == null) return defaultValue;
        try {
            return new BigDecimal(val.toString());
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }

    private int parseInteger(Map<String, Object> map, String key, int defaultValue) {
        Object val = map.get(key);
        if (val == null) return defaultValue;
        try {
            return Integer.parseInt(val.toString());
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }

    private Instant parseInstant(Map<String, Object> map, String key) {
        Object val = map.get(key);
        if (val == null) return Instant.now();
        try {
            return Instant.parse(val.toString());
        } catch (Exception e) {
            return Instant.now();
        }
    }
}
