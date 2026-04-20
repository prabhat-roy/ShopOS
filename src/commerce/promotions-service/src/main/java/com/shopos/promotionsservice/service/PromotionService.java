package com.shopos.promotionsservice.service;

import com.shopos.promotionsservice.domain.Promotion;
import com.shopos.promotionsservice.domain.PromoType;
import com.shopos.promotionsservice.dto.CreatePromotionRequest;
import com.shopos.promotionsservice.dto.ValidateRequest;
import com.shopos.promotionsservice.dto.ValidateResponse;
import com.shopos.promotionsservice.repository.PromotionRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.Instant;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class PromotionService {

    private final PromotionRepository promotionRepository;

    // ── CRUD ──────────────────────────────────────────────────────────────────

    @Transactional
    public Promotion createPromotion(CreatePromotionRequest req) {
        if (promotionRepository.findByCode(req.code()).isPresent()) {
            throw new IllegalArgumentException("Promotion code already exists: " + req.code());
        }

        Promotion promo = Promotion.builder()
                .code(req.code().toUpperCase().strip())
                .name(req.name())
                .type(req.type())
                .discountValue(req.discountValue())
                .discountPercent(req.discountPercent())
                .minOrderAmount(req.minOrderAmount() != null ? req.minOrderAmount() : BigDecimal.ZERO)
                .maxUses(req.maxUses() != null ? req.maxUses() : 0)
                .active(true)
                .startsAt(req.startsAt())
                .expiresAt(req.expiresAt())
                .build();

        promo = promotionRepository.save(promo);
        log.info("Promotion created id={} code={}", promo.getId(), promo.getCode());
        return promo;
    }

    @Transactional(readOnly = true)
    public Promotion getPromotion(UUID id) {
        return findOrThrow(id);
    }

    @Transactional(readOnly = true)
    public List<Promotion> listPromotions(boolean activeOnly) {
        if (activeOnly) {
            Instant now = Instant.now();
            return promotionRepository.findByActiveTrueAndStartsAtBeforeAndExpiresAtAfter(now, now);
        }
        return promotionRepository.findAll();
    }

    @Transactional
    public void deactivate(UUID id) {
        Promotion promo = findOrThrow(id);
        promo.setActive(false);
        promotionRepository.save(promo);
        log.info("Promotion deactivated id={}", id);
    }

    // ── Validate / Apply ──────────────────────────────────────────────────────

    /**
     * Validates a coupon code against business rules without incrementing usedCount.
     * Returns a ValidateResponse with valid=true and the computed discount, or
     * valid=false with a human-readable reason.
     */
    @Transactional(readOnly = true)
    public ValidateResponse validateCoupon(ValidateRequest req) {
        String upperCode = req.code().toUpperCase().strip();

        // 1. Code must exist
        Promotion promo = promotionRepository.findByCode(upperCode).orElse(null);
        if (promo == null) {
            return ValidateResponse.fail(upperCode, "Promotion code not found");
        }

        // 2. Must be active
        if (!Boolean.TRUE.equals(promo.getActive())) {
            return ValidateResponse.fail(upperCode, "Promotion is not active");
        }

        // 3. Validity window
        Instant now = Instant.now();
        if (promo.getStartsAt() != null && now.isBefore(promo.getStartsAt())) {
            return ValidateResponse.fail(upperCode, "Promotion has not started yet");
        }
        if (promo.getExpiresAt() != null && now.isAfter(promo.getExpiresAt())) {
            return ValidateResponse.fail(upperCode, "Promotion has expired");
        }

        // 4. Max uses (0 = unlimited)
        if (promo.getMaxUses() > 0 && promo.getUsedCount() >= promo.getMaxUses()) {
            return ValidateResponse.fail(upperCode, "Promotion has reached its maximum uses");
        }

        // 5. Minimum order amount
        if (promo.getMinOrderAmount() != null
                && req.orderAmount().compareTo(promo.getMinOrderAmount()) < 0) {
            return ValidateResponse.fail(upperCode,
                    "Order amount " + req.orderAmount()
                    + " is below minimum required " + promo.getMinOrderAmount());
        }

        BigDecimal discount = computeDiscount(promo, req.orderAmount());
        return ValidateResponse.ok(upperCode, discount);
    }

    /**
     * Increments usedCount and returns the computed discount amount.
     * Throws IllegalArgumentException if the code does not exist.
     */
    @Transactional
    public BigDecimal applyCoupon(String code) {
        String upperCode = code.toUpperCase().strip();
        Promotion promo = promotionRepository.findByCode(upperCode)
                .orElseThrow(() -> new IllegalArgumentException(
                        "Promotion code not found: " + upperCode));

        promo.setUsedCount(promo.getUsedCount() + 1);
        promotionRepository.save(promo);
        log.info("Promotion applied code={} usedCount={}", upperCode, promo.getUsedCount());

        // Return discount for the apply response (no order amount context here — return 0)
        return promo.getDiscountValue() != null ? promo.getDiscountValue() : BigDecimal.ZERO;
    }

    // ── helpers ───────────────────────────────────────────────────────────────

    private Promotion findOrThrow(UUID id) {
        return promotionRepository.findById(id)
                .orElseThrow(() -> new NoSuchElementException("Promotion not found: " + id));
    }

    /**
     * Computes the discount amount for a given promotion and order amount.
     * FREE_SHIPPING and BUY_X_GET_Y return 0 here (fulfillment logic lives
     * in checkout-service); the shipping/item credit is applied separately.
     */
    private BigDecimal computeDiscount(Promotion promo, BigDecimal orderAmount) {
        return switch (promo.getType()) {
            case PERCENTAGE -> {
                if (promo.getDiscountPercent() == null) yield BigDecimal.ZERO;
                yield orderAmount
                        .multiply(promo.getDiscountPercent())
                        .divide(BigDecimal.valueOf(100), 2, RoundingMode.HALF_UP);
            }
            case FIXED_AMOUNT -> {
                BigDecimal val = promo.getDiscountValue() != null
                        ? promo.getDiscountValue() : BigDecimal.ZERO;
                // Discount cannot exceed the order amount
                yield val.min(orderAmount);
            }
            case FREE_SHIPPING, BUY_X_GET_Y -> BigDecimal.ZERO;
        };
    }
}
