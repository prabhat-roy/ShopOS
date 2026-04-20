package com.shopos.promotionsservice.service;

import com.shopos.promotionsservice.domain.Promotion;
import com.shopos.promotionsservice.domain.PromoType;
import com.shopos.promotionsservice.dto.ValidateRequest;
import com.shopos.promotionsservice.dto.ValidateResponse;
import com.shopos.promotionsservice.repository.PromotionRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class PromotionServiceTest {

    @Mock
    private PromotionRepository promotionRepository;

    @InjectMocks
    private PromotionService promotionService;

    private Promotion activePromo;

    @BeforeEach
    void setUp() {
        activePromo = Promotion.builder()
                .id(UUID.randomUUID())
                .code("SAVE20")
                .name("Save 20%")
                .type(PromoType.PERCENTAGE)
                .discountPercent(new BigDecimal("20.00"))
                .minOrderAmount(new BigDecimal("50.00"))
                .maxUses(100)
                .usedCount(10)
                .active(true)
                .startsAt(Instant.now().minus(1, ChronoUnit.DAYS))
                .expiresAt(Instant.now().plus(7, ChronoUnit.DAYS))
                .build();
    }

    // ── validateCoupon: valid code ────────────────────────────────────────────

    @Test
    @DisplayName("validateCoupon: returns valid=true and correct discount for a valid PERCENTAGE code")
    void validateCoupon_validCode_returnsValidResponse() {
        when(promotionRepository.findByCode("SAVE20")).thenReturn(Optional.of(activePromo));

        ValidateRequest req = new ValidateRequest("SAVE20", new BigDecimal("100.00"), "cust-001");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isTrue();
        assertThat(response.code()).isEqualTo("SAVE20");
        // 20% of 100 = 20.00
        assertThat(response.discountAmount()).isEqualByComparingTo(new BigDecimal("20.00"));
        assertThat(response.reason()).isNull();
    }

    @Test
    @DisplayName("validateCoupon: returns valid=true for a FIXED_AMOUNT code")
    void validateCoupon_fixedAmount_returnsCorrectDiscount() {
        Promotion fixedPromo = Promotion.builder()
                .id(UUID.randomUUID())
                .code("FLAT10")
                .name("$10 Off")
                .type(PromoType.FIXED_AMOUNT)
                .discountValue(new BigDecimal("10.00"))
                .minOrderAmount(new BigDecimal("30.00"))
                .maxUses(0) // unlimited
                .usedCount(500)
                .active(true)
                .startsAt(Instant.now().minus(1, ChronoUnit.DAYS))
                .expiresAt(Instant.now().plus(30, ChronoUnit.DAYS))
                .build();

        when(promotionRepository.findByCode("FLAT10")).thenReturn(Optional.of(fixedPromo));

        ValidateRequest req = new ValidateRequest("FLAT10", new BigDecimal("80.00"), "cust-002");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isTrue();
        assertThat(response.discountAmount()).isEqualByComparingTo(new BigDecimal("10.00"));
    }

    // ── validateCoupon: code not found ────────────────────────────────────────

    @Test
    @DisplayName("validateCoupon: returns valid=false when code does not exist")
    void validateCoupon_unknownCode_returnsInvalid() {
        when(promotionRepository.findByCode("BOGUS")).thenReturn(Optional.empty());

        ValidateRequest req = new ValidateRequest("BOGUS", new BigDecimal("100.00"), "cust-001");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isFalse();
        assertThat(response.reason()).contains("not found");
    }

    // ── validateCoupon: expired ───────────────────────────────────────────────

    @Test
    @DisplayName("validateCoupon: returns valid=false when promotion has expired")
    void validateCoupon_expiredPromo_returnsInvalid() {
        Promotion expired = Promotion.builder()
                .id(UUID.randomUUID())
                .code("OLD10")
                .name("Old deal")
                .type(PromoType.PERCENTAGE)
                .discountPercent(new BigDecimal("10.00"))
                .minOrderAmount(BigDecimal.ZERO)
                .maxUses(0)
                .usedCount(0)
                .active(true)
                .startsAt(Instant.now().minus(10, ChronoUnit.DAYS))
                .expiresAt(Instant.now().minus(1, ChronoUnit.DAYS))   // already expired
                .build();

        when(promotionRepository.findByCode("OLD10")).thenReturn(Optional.of(expired));

        ValidateRequest req = new ValidateRequest("OLD10", new BigDecimal("100.00"), "cust-001");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isFalse();
        assertThat(response.reason()).containsIgnoringCase("expired");
    }

    // ── validateCoupon: max uses exceeded ────────────────────────────────────

    @Test
    @DisplayName("validateCoupon: returns valid=false when max uses have been reached")
    void validateCoupon_maxUsesExceeded_returnsInvalid() {
        Promotion maxedOut = Promotion.builder()
                .id(UUID.randomUUID())
                .code("MAXED")
                .name("Sold out promo")
                .type(PromoType.PERCENTAGE)
                .discountPercent(new BigDecimal("15.00"))
                .minOrderAmount(BigDecimal.ZERO)
                .maxUses(50)
                .usedCount(50)    // fully exhausted
                .active(true)
                .startsAt(Instant.now().minus(1, ChronoUnit.DAYS))
                .expiresAt(Instant.now().plus(1, ChronoUnit.DAYS))
                .build();

        when(promotionRepository.findByCode("MAXED")).thenReturn(Optional.of(maxedOut));

        ValidateRequest req = new ValidateRequest("MAXED", new BigDecimal("100.00"), "cust-001");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isFalse();
        assertThat(response.reason()).containsIgnoringCase("maximum uses");
    }

    // ── validateCoupon: below minimum order amount ────────────────────────────

    @Test
    @DisplayName("validateCoupon: returns valid=false when order amount is below minimum")
    void validateCoupon_belowMinOrderAmount_returnsInvalid() {
        // activePromo requires min 50.00 — send 30.00
        when(promotionRepository.findByCode("SAVE20")).thenReturn(Optional.of(activePromo));

        ValidateRequest req = new ValidateRequest("SAVE20", new BigDecimal("30.00"), "cust-001");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isFalse();
        assertThat(response.reason()).containsIgnoringCase("minimum");
    }

    // ── validateCoupon: inactive promotion ───────────────────────────────────

    @Test
    @DisplayName("validateCoupon: returns valid=false when promotion is inactive")
    void validateCoupon_inactivePromo_returnsInvalid() {
        Promotion inactive = Promotion.builder()
                .id(UUID.randomUUID())
                .code("DEAD50")
                .name("Disabled promo")
                .type(PromoType.FIXED_AMOUNT)
                .discountValue(new BigDecimal("50.00"))
                .minOrderAmount(BigDecimal.ZERO)
                .maxUses(0)
                .usedCount(0)
                .active(false)
                .startsAt(Instant.now().minus(1, ChronoUnit.DAYS))
                .expiresAt(Instant.now().plus(30, ChronoUnit.DAYS))
                .build();

        when(promotionRepository.findByCode("DEAD50")).thenReturn(Optional.of(inactive));

        ValidateRequest req = new ValidateRequest("DEAD50", new BigDecimal("200.00"), "cust-001");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isFalse();
        assertThat(response.reason()).containsIgnoringCase("not active");
    }

    // ── validateCoupon: FREE_SHIPPING returns zero discount ──────────────────

    @Test
    @DisplayName("validateCoupon: FREE_SHIPPING promo returns valid=true with discountAmount=0")
    void validateCoupon_freeShipping_returnsZeroDiscount() {
        Promotion freeShip = Promotion.builder()
                .id(UUID.randomUUID())
                .code("FREESHIP")
                .name("Free Shipping")
                .type(PromoType.FREE_SHIPPING)
                .minOrderAmount(BigDecimal.ZERO)
                .maxUses(0)
                .usedCount(0)
                .active(true)
                .startsAt(Instant.now().minus(1, ChronoUnit.DAYS))
                .expiresAt(Instant.now().plus(7, ChronoUnit.DAYS))
                .build();

        when(promotionRepository.findByCode("FREESHIP")).thenReturn(Optional.of(freeShip));

        ValidateRequest req = new ValidateRequest("FREESHIP", new BigDecimal("45.00"), "cust-003");
        ValidateResponse response = promotionService.validateCoupon(req);

        assertThat(response.valid()).isTrue();
        assertThat(response.discountAmount()).isEqualByComparingTo(BigDecimal.ZERO);
    }
}
