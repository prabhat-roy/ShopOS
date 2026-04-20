package com.shopos.marketplacesellerservice.service;

import com.shopos.marketplacesellerservice.domain.Seller;
import com.shopos.marketplacesellerservice.domain.SellerProduct;
import com.shopos.marketplacesellerservice.domain.SellerStatus;
import com.shopos.marketplacesellerservice.domain.SellerTier;
import com.shopos.marketplacesellerservice.dto.CreateSellerRequest;
import com.shopos.marketplacesellerservice.dto.ListingRequest;
import com.shopos.marketplacesellerservice.dto.SellerResponse;
import com.shopos.marketplacesellerservice.repository.SellerProductRepository;
import com.shopos.marketplacesellerservice.repository.SellerRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import java.util.UUID;

/**
 * Core business logic for marketplace seller management.
 */
@Slf4j
@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class MarketplaceSellerService {

    private final SellerRepository sellerRepository;
    private final SellerProductRepository sellerProductRepository;

    // -------------------------------------------------------------------------
    // Seller onboarding & lifecycle
    // -------------------------------------------------------------------------

    /**
     * Registers a new seller in PENDING status awaiting admin approval.
     *
     * @throws IllegalStateException when a seller already exists for the given orgId
     */
    @Transactional
    public SellerResponse onboardSeller(CreateSellerRequest request) {
        sellerRepository.findByOrgId(request.orgId()).ifPresent(s -> {
            throw new IllegalStateException(
                    "A seller already exists for organisation: " + request.orgId());
        });

        BigDecimal commissionRate = request.commissionRate() != null
                ? request.commissionRate()
                : new BigDecimal("15.00");

        Seller seller = Seller.builder()
                .orgId(request.orgId())
                .displayName(request.displayName())
                .description(request.description())
                .status(SellerStatus.PENDING)
                .tier(SellerTier.BRONZE)
                .commissionRate(commissionRate)
                .build();

        Seller saved = sellerRepository.save(seller);
        log.info("Seller onboarded — id={}, org={}", saved.getId(), saved.getOrgId());
        return SellerResponse.from(saved);
    }

    /**
     * Transitions a seller from PENDING to ACTIVE.
     *
     * @throws IllegalStateException when the seller is not in PENDING status
     */
    @Transactional
    public SellerResponse approveSeller(UUID sellerId) {
        Seller seller = requireSeller(sellerId);

        if (seller.getStatus() != SellerStatus.PENDING) {
            throw new IllegalStateException(
                    "Seller must be in PENDING status to approve. Current status: " + seller.getStatus());
        }

        seller.setStatus(SellerStatus.ACTIVE);
        seller.setOnboardedAt(Instant.now());
        Seller saved = sellerRepository.save(seller);
        log.info("Seller approved — id={}", sellerId);
        return SellerResponse.from(saved);
    }

    /**
     * Suspends an ACTIVE seller. Listings remain but are hidden from buyers.
     *
     * @throws IllegalStateException when the seller is not ACTIVE
     */
    @Transactional
    public SellerResponse suspendSeller(UUID sellerId) {
        Seller seller = requireSeller(sellerId);

        if (seller.getStatus() != SellerStatus.ACTIVE) {
            throw new IllegalStateException(
                    "Only ACTIVE sellers can be suspended. Current status: " + seller.getStatus());
        }

        seller.setStatus(SellerStatus.SUSPENDED);
        Seller saved = sellerRepository.save(seller);
        log.info("Seller suspended — id={}", sellerId);
        return SellerResponse.from(saved);
    }

    /**
     * Permanently terminates a seller. This is irreversible.
     *
     * @throws IllegalStateException when the seller is already TERMINATED
     */
    @Transactional
    public SellerResponse terminateSeller(UUID sellerId) {
        Seller seller = requireSeller(sellerId);

        if (seller.getStatus() == SellerStatus.TERMINATED) {
            throw new IllegalStateException("Seller is already terminated.");
        }

        seller.setStatus(SellerStatus.TERMINATED);
        Seller saved = sellerRepository.save(seller);
        log.info("Seller terminated — id={}", sellerId);
        return SellerResponse.from(saved);
    }

    // -------------------------------------------------------------------------
    // Seller reads
    // -------------------------------------------------------------------------

    public SellerResponse getSeller(UUID sellerId) {
        return SellerResponse.from(requireSeller(sellerId));
    }

    public SellerResponse getByOrg(UUID orgId) {
        Seller seller = sellerRepository.findByOrgId(orgId)
                .orElseThrow(() -> new IllegalArgumentException(
                        "No seller found for organisation: " + orgId));
        return SellerResponse.from(seller);
    }

    /**
     * Lists sellers, optionally filtered by status and/or tier.
     */
    public List<SellerResponse> listSellers(SellerStatus status, SellerTier tier) {
        List<Seller> sellers;
        if (status != null && tier != null) {
            sellers = sellerRepository.findByStatusAndTier(status, tier);
        } else if (status != null) {
            sellers = sellerRepository.findByStatus(status);
        } else if (tier != null) {
            sellers = sellerRepository.findByTier(tier);
        } else {
            sellers = sellerRepository.findAll();
        }
        return sellers.stream().map(SellerResponse::from).toList();
    }

    // -------------------------------------------------------------------------
    // Performance stats & tier recalculation
    // -------------------------------------------------------------------------

    /**
     * Updates cumulative performance statistics for a seller and recalculates tier.
     *
     * <p>Tier thresholds:
     * <ul>
     *   <li>BRONZE:   0–49 total orders</li>
     *   <li>SILVER:   50–199 total orders</li>
     *   <li>GOLD:     200–999 total orders</li>
     *   <li>PLATINUM: 1000+ total orders</li>
     * </ul>
     *
     * @param id          seller UUID
     * @param ordersCount additional orders to add to the running total
     * @param salesAmount additional revenue to add to the running total
     * @param returnRate  new return rate percentage (replaces previous value)
     */
    @Transactional
    public SellerResponse updateSellerStats(UUID id, int ordersCount,
                                           BigDecimal salesAmount, BigDecimal returnRate) {
        Seller seller = requireSeller(id);

        int newTotalOrders = seller.getTotalOrders() + ordersCount;
        BigDecimal newTotalSales = seller.getTotalSales().add(salesAmount);

        seller.setTotalOrders(newTotalOrders);
        seller.setTotalSales(newTotalSales);
        seller.setReturnRate(returnRate);
        seller.setTier(calculateTier(newTotalOrders));

        Seller saved = sellerRepository.save(seller);
        log.info("Seller stats updated — id={}, totalOrders={}, tier={}",
                id, newTotalOrders, saved.getTier());
        return SellerResponse.from(saved);
    }

    // -------------------------------------------------------------------------
    // Product listings
    // -------------------------------------------------------------------------

    /**
     * Creates a new product listing for the seller.
     *
     * @throws IllegalStateException when the seller is not ACTIVE
     * @throws IllegalStateException when a listing with the same sellerSku already exists
     */
    @Transactional
    public SellerProduct createListing(UUID sellerId, ListingRequest request) {
        Seller seller = requireSeller(sellerId);

        if (seller.getStatus() != SellerStatus.ACTIVE) {
            throw new IllegalStateException(
                    "Only ACTIVE sellers can create listings. Status: " + seller.getStatus());
        }

        sellerProductRepository.findBySellerIdAndSellerSku(sellerId, request.sellerSku())
                .ifPresent(p -> {
                    throw new IllegalStateException(
                            "A listing with sellerSku '" + request.sellerSku() + "' already exists");
                });

        SellerProduct product = SellerProduct.builder()
                .sellerId(sellerId)
                .productId(request.productId())
                .sku(request.sku())
                .sellerSku(request.sellerSku())
                .listingPrice(request.listingPrice())
                .stockQuantity(request.stockQuantity())
                .status("PENDING")
                .build();

        SellerProduct saved = sellerProductRepository.save(product);
        seller.setProductCount(seller.getProductCount() + 1);
        sellerRepository.save(seller);

        log.info("Listing created — seller={}, listingId={}", sellerId, saved.getId());
        return saved;
    }

    public SellerProduct getListing(UUID listingId) {
        return sellerProductRepository.findById(listingId)
                .orElseThrow(() -> new IllegalArgumentException(
                        "Listing not found: " + listingId));
    }

    /**
     * Returns all listings for a seller, optionally filtered by status.
     */
    public List<SellerProduct> listSellerProducts(UUID sellerId, String status) {
        if (status != null && !status.isBlank()) {
            return sellerProductRepository.findBySellerIdAndStatus(sellerId, status.toUpperCase());
        }
        return sellerProductRepository.findBySellerId(sellerId);
    }

    /**
     * Updates price and stock for an existing listing.
     */
    @Transactional
    public SellerProduct updateListing(UUID listingId, ListingRequest request) {
        SellerProduct product = getListing(listingId);
        product.setListingPrice(request.listingPrice());
        product.setStockQuantity(request.stockQuantity());
        product.setSku(request.sku());
        product.setSellerSku(request.sellerSku());
        return sellerProductRepository.save(product);
    }

    /**
     * Sets a listing status to INACTIVE, hiding it from buyers.
     */
    @Transactional
    public SellerProduct deactivateListing(UUID listingId) {
        SellerProduct product = getListing(listingId);
        if ("INACTIVE".equals(product.getStatus())) {
            throw new IllegalStateException("Listing is already inactive");
        }
        product.setStatus("INACTIVE");
        return sellerProductRepository.save(product);
    }

    // -------------------------------------------------------------------------
    // private helpers
    // -------------------------------------------------------------------------

    private Seller requireSeller(UUID sellerId) {
        return sellerRepository.findById(sellerId)
                .orElseThrow(() -> new IllegalArgumentException(
                        "Seller not found: " + sellerId));
    }

    /**
     * Derives performance tier from the cumulative order count.
     */
    private SellerTier calculateTier(int totalOrders) {
        if (totalOrders >= 1000) {
            return SellerTier.PLATINUM;
        } else if (totalOrders >= 200) {
            return SellerTier.GOLD;
        } else if (totalOrders >= 50) {
            return SellerTier.SILVER;
        } else {
            return SellerTier.BRONZE;
        }
    }
}
