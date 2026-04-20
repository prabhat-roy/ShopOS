package com.shopos.pricingservice;

import com.shopos.pricingservice.domain.Price;
import com.shopos.pricingservice.domain.PriceCalculation;
import com.shopos.pricingservice.dto.CalculateRequest;
import com.shopos.pricingservice.repository.PriceRepository;
import com.shopos.pricingservice.service.PricingService;
import jakarta.persistence.EntityNotFoundException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class PricingServiceTest {

    @Mock
    private PriceRepository priceRepository;

    @InjectMocks
    private PricingService pricingService;

    private Price basePriceTier1;
    private Price basePriceTier10;

    @BeforeEach
    void setUp() {
        // Tier 1: min qty 1, base price 100.00, no sale
        basePriceTier1 = Price.builder()
                .productId("prod-001")
                .currency("USD")
                .basePrice(new BigDecimal("100.00"))
                .salePrice(null)
                .minQty(1)
                .segment("all")
                .active(true)
                .build();

        // Tier 10: min qty 10, base price 90.00, no sale — bulk discount
        basePriceTier10 = Price.builder()
                .productId("prod-001")
                .currency("USD")
                .basePrice(new BigDecimal("90.00"))
                .salePrice(null)
                .minQty(10)
                .segment("all")
                .active(true)
                .build();
    }

    @Test
    @DisplayName("calculate returns base price for quantity below bulk threshold")
    void calculate_singleUnit_returnsBasePrice() {
        // Quantity 1 — only tier-1 qualifies
        when(priceRepository.findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                eq("prod-001"), eq("all"), eq(1)))
                .thenReturn(List.of(basePriceTier1));

        CalculateRequest req = new CalculateRequest("prod-001", 1, "all");
        PriceCalculation result = pricingService.calculate(req);

        assertThat(result.finalPrice()).isEqualByComparingTo("100.00");
        assertThat(result.discountPercent()).isEqualByComparingTo("0.00");
        assertThat(result.currency()).isEqualTo("USD");
    }

    @Test
    @DisplayName("calculate returns tiered bulk price when quantity meets threshold")
    void calculate_bulkQuantity_returnsTieredPrice() {
        // Quantity 15 — both tiers qualify; tier-10 (higher min_qty) is returned first by the query
        when(priceRepository.findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                eq("prod-001"), eq("all"), eq(15)))
                .thenReturn(List.of(basePriceTier10, basePriceTier1));

        CalculateRequest req = new CalculateRequest("prod-001", 15, "all");
        PriceCalculation result = pricingService.calculate(req);

        assertThat(result.finalPrice()).isEqualByComparingTo("90.00");
        assertThat(result.basePrice()).isEqualByComparingTo("90.00");
    }

    @Test
    @DisplayName("calculate applies sale price when sale is active and within time window")
    void calculate_activeSalePrice_appliesSale() {
        Price salePrice = Price.builder()
                .productId("prod-002")
                .currency("USD")
                .basePrice(new BigDecimal("80.00"))
                .salePrice(new BigDecimal("60.00"))
                .minQty(1)
                .segment("all")
                .active(true)
                .startAt(OffsetDateTime.now().minusDays(1))
                .endAt(OffsetDateTime.now().plusDays(1))
                .build();

        when(priceRepository.findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                eq("prod-002"), eq("all"), eq(1)))
                .thenReturn(List.of(salePrice));

        CalculateRequest req = new CalculateRequest("prod-002", 1, "all");
        PriceCalculation result = pricingService.calculate(req);

        assertThat(result.finalPrice()).isEqualByComparingTo("60.00");
        assertThat(result.basePrice()).isEqualByComparingTo("80.00");
        // Discount = (80 - 60) / 80 * 100 = 25%
        assertThat(result.discountPercent()).isEqualByComparingTo("25.00");
    }

    @Test
    @DisplayName("calculate does NOT apply sale price when outside time window")
    void calculate_expiredSalePrice_returnsBasePrice() {
        Price expiredSale = Price.builder()
                .productId("prod-003")
                .currency("USD")
                .basePrice(new BigDecimal("50.00"))
                .salePrice(new BigDecimal("40.00"))
                .minQty(1)
                .segment("all")
                .active(true)
                .startAt(OffsetDateTime.now().minusDays(10))
                .endAt(OffsetDateTime.now().minusDays(1))   // already expired
                .build();

        when(priceRepository.findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                eq("prod-003"), eq("all"), eq(1)))
                .thenReturn(List.of(expiredSale));

        CalculateRequest req = new CalculateRequest("prod-003", 1, "all");
        PriceCalculation result = pricingService.calculate(req);

        assertThat(result.finalPrice()).isEqualByComparingTo("50.00");
        assertThat(result.discountPercent()).isEqualByComparingTo("0.00");
    }

    @Test
    @DisplayName("calculate falls back to 'all' segment when no segment-specific tier exists")
    void calculate_segmentSpecific_fallsBackToAll() {
        // No VIP-specific tier for qty 1
        when(priceRepository.findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                eq("prod-001"), eq("vip"), eq(1)))
                .thenReturn(List.of());

        when(priceRepository.findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                eq("prod-001"), eq("all"), eq(1)))
                .thenReturn(List.of(basePriceTier1));

        CalculateRequest req = new CalculateRequest("prod-001", 1, "vip");
        PriceCalculation result = pricingService.calculate(req);

        assertThat(result.finalPrice()).isEqualByComparingTo("100.00");
    }

    @Test
    @DisplayName("calculate throws EntityNotFoundException when no price exists")
    void calculate_noPrice_throwsEntityNotFoundException() {
        when(priceRepository.findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                anyString(), anyString(), anyInt()))
                .thenReturn(List.of());

        CalculateRequest req = new CalculateRequest("prod-unknown", 1, "all");

        assertThatThrownBy(() -> pricingService.calculate(req))
                .isInstanceOf(EntityNotFoundException.class)
                .hasMessageContaining("prod-unknown");
    }
}
