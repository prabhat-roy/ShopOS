package com.shopos.pricingservice.service;

import com.shopos.pricingservice.domain.Price;
import com.shopos.pricingservice.domain.PriceCalculation;
import com.shopos.pricingservice.dto.CalculateRequest;
import com.shopos.pricingservice.dto.SetPriceRequest;
import com.shopos.pricingservice.repository.PriceRepository;
import jakarta.persistence.EntityNotFoundException;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class PricingService {

    private final PriceRepository priceRepository;

    /**
     * Upserts a price record. If an active record already exists for the same
     * (productId, currency, segment, minQty) combination, it is updated in place;
     * otherwise a new record is created.
     */
    @Transactional
    public Price setPrice(SetPriceRequest req) {
        // Try to find an existing active record with matching composite key
        List<Price> existing = priceRepository
                .findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                        req.productId(), req.segment(), req.minQty());

        Price price = existing.stream()
                .filter(p -> p.getMinQty() == req.minQty()
                        && p.getCurrency().equalsIgnoreCase(req.currency()))
                .findFirst()
                .orElse(null);

        if (price == null) {
            price = Price.builder()
                    .productId(req.productId())
                    .currency(req.currency())
                    .segment(req.segment())
                    .minQty(req.minQty())
                    .build();
        }

        price.setBasePrice(req.basePrice());
        price.setSalePrice(req.salePrice());
        price.setStartAt(req.startAt());
        price.setEndAt(req.endAt());
        price.setActive(true);

        return priceRepository.save(price);
    }

    /**
     * Returns all active price records for a product in the given currency across all segments.
     */
    @Transactional(readOnly = true)
    public List<Price> getPrice(String productId, String currency) {
        return priceRepository.findByProductIdAndActiveTrue(productId)
                .stream()
                .filter(p -> p.getCurrency().equalsIgnoreCase(currency))
                .toList();
    }

    /**
     * Calculates the effective price for a product given quantity and customer segment.
     * <p>
     * Resolution order:
     * 1. Look for the best tiered price matching the customer segment exactly.
     * 2. Fall back to the "all" segment if no segment-specific tier matches.
     * 3. Apply sale price if it is set and the current time is within [startAt, endAt].
     */
    @Transactional(readOnly = true)
    public PriceCalculation calculate(CalculateRequest req) {
        Price best = resolveBestPrice(req.productId(), req.quantity(), req.segment());

        if (best == null) {
            throw new EntityNotFoundException(
                    "No active price found for productId=" + req.productId()
                            + " segment=" + req.segment()
                            + " quantity=" + req.quantity());
        }

        BigDecimal finalPrice = computeFinalPrice(best);

        BigDecimal discountPercent = BigDecimal.ZERO;
        if (finalPrice.compareTo(best.getBasePrice()) < 0) {
            discountPercent = best.getBasePrice()
                    .subtract(finalPrice)
                    .divide(best.getBasePrice(), 4, RoundingMode.HALF_UP)
                    .multiply(BigDecimal.valueOf(100))
                    .setScale(2, RoundingMode.HALF_UP);
        }

        return new PriceCalculation(
                req.productId(),
                req.quantity(),
                req.segment(),
                best.getBasePrice(),
                finalPrice,
                discountPercent,
                best.getCurrency()
        );
    }

    /**
     * Returns all price records (active and inactive) for a product.
     */
    @Transactional(readOnly = true)
    public List<Price> listPrices(String productId) {
        return priceRepository.findAll().stream()
                .filter(p -> p.getProductId().equals(productId))
                .toList();
    }

    /**
     * Deactivates a price record (soft delete).
     */
    @Transactional
    public void deletePrice(UUID id) {
        Price price = priceRepository.findById(id)
                .orElseThrow(() -> new EntityNotFoundException("Price not found with id=" + id));
        price.setActive(false);
        priceRepository.save(price);
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    private Price resolveBestPrice(String productId, int quantity, String segment) {
        // Try specific segment first
        if (!"all".equalsIgnoreCase(segment)) {
            List<Price> segmentTiers = priceRepository
                    .findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                            productId, segment, quantity);
            Price candidate = segmentTiers.stream()
                    .filter(Price::isActive)
                    .findFirst()
                    .orElse(null);
            if (candidate != null) return candidate;
        }

        // Fall back to "all" segment
        List<Price> allTiers = priceRepository
                .findByProductIdAndSegmentAndMinQtyLessThanEqualOrderByMinQtyDesc(
                        productId, "all", quantity);
        return allTiers.stream()
                .filter(Price::isActive)
                .findFirst()
                .orElse(null);
    }

    private BigDecimal computeFinalPrice(Price price) {
        if (price.getSalePrice() == null) {
            return price.getBasePrice();
        }

        OffsetDateTime now = OffsetDateTime.now();
        boolean withinStart = price.getStartAt() == null || !now.isBefore(price.getStartAt());
        boolean withinEnd   = price.getEndAt()   == null || !now.isAfter(price.getEndAt());

        if (withinStart && withinEnd) {
            return price.getSalePrice();
        }

        return price.getBasePrice();
    }
}
