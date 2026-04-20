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
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class MarketplaceSellerServiceTest {

    @Mock
    private SellerRepository sellerRepository;

    @Mock
    private SellerProductRepository sellerProductRepository;

    @InjectMocks
    private MarketplaceSellerService sellerService;

    private UUID sellerId;
    private UUID orgId;
    private Seller pendingSeller;
    private Seller activeSeller;

    @BeforeEach
    void setUp() {
        sellerId = UUID.randomUUID();
        orgId = UUID.randomUUID();

        pendingSeller = Seller.builder()
                .id(sellerId)
                .orgId(orgId)
                .displayName("Test Seller Ltd")
                .description("A test marketplace seller")
                .status(SellerStatus.PENDING)
                .tier(SellerTier.BRONZE)
                .commissionRate(new BigDecimal("15.00"))
                .rating(BigDecimal.ZERO)
                .totalSales(BigDecimal.ZERO)
                .totalOrders(0)
                .productCount(0)
                .returnRate(BigDecimal.ZERO)
                .build();

        activeSeller = Seller.builder()
                .id(sellerId)
                .orgId(orgId)
                .displayName("Test Seller Ltd")
                .status(SellerStatus.ACTIVE)
                .tier(SellerTier.BRONZE)
                .commissionRate(new BigDecimal("15.00"))
                .rating(BigDecimal.ZERO)
                .totalSales(BigDecimal.ZERO)
                .totalOrders(0)
                .productCount(0)
                .returnRate(BigDecimal.ZERO)
                .onboardedAt(Instant.now())
                .build();
    }

    // -------------------------------------------------------------------------
    // Test 1: Onboard new seller — creates in PENDING status
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("1 — onboardSeller creates seller with PENDING status")
    void onboardSeller_newOrg_createsPendingSeller() {
        CreateSellerRequest request = new CreateSellerRequest(
                orgId, "Test Seller Ltd", "Description", new BigDecimal("12.00"));

        when(sellerRepository.findByOrgId(orgId)).thenReturn(Optional.empty());
        when(sellerRepository.save(any(Seller.class))).thenReturn(pendingSeller);

        SellerResponse response = sellerService.onboardSeller(request);

        assertThat(response.status()).isEqualTo(SellerStatus.PENDING);
        assertThat(response.displayName()).isEqualTo("Test Seller Ltd");
        verify(sellerRepository).save(any(Seller.class));
    }

    // -------------------------------------------------------------------------
    // Test 2: Onboard duplicate org — throws IllegalStateException
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("2 — onboardSeller with existing org throws IllegalStateException")
    void onboardSeller_duplicateOrg_throwsIllegalStateException() {
        CreateSellerRequest request = new CreateSellerRequest(
                orgId, "Duplicate", null, null);

        when(sellerRepository.findByOrgId(orgId)).thenReturn(Optional.of(pendingSeller));

        assertThatThrownBy(() -> sellerService.onboardSeller(request))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining(orgId.toString());
    }

    // -------------------------------------------------------------------------
    // Test 3: Approve PENDING seller transitions to ACTIVE
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("3 — approveSeller transitions PENDING seller to ACTIVE and sets onboardedAt")
    void approveSeller_pendingSeller_transitionsToActive() {
        when(sellerRepository.findById(sellerId)).thenReturn(Optional.of(pendingSeller));
        when(sellerRepository.save(any(Seller.class))).thenAnswer(inv -> inv.getArgument(0));

        SellerResponse response = sellerService.approveSeller(sellerId);

        assertThat(response.status()).isEqualTo(SellerStatus.ACTIVE);
        assertThat(response.onboardedAt()).isNotNull();
    }

    // -------------------------------------------------------------------------
    // Test 4: Approve non-PENDING seller throws IllegalStateException
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("4 — approveSeller on ACTIVE seller throws IllegalStateException")
    void approveSeller_activeSeller_throwsIllegalStateException() {
        when(sellerRepository.findById(sellerId)).thenReturn(Optional.of(activeSeller));

        assertThatThrownBy(() -> sellerService.approveSeller(sellerId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("PENDING");
    }

    // -------------------------------------------------------------------------
    // Test 5: Suspend ACTIVE seller
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("5 — suspendSeller transitions ACTIVE seller to SUSPENDED")
    void suspendSeller_activeSeller_transitionsToSuspended() {
        when(sellerRepository.findById(sellerId)).thenReturn(Optional.of(activeSeller));
        when(sellerRepository.save(any(Seller.class))).thenAnswer(inv -> inv.getArgument(0));

        SellerResponse response = sellerService.suspendSeller(sellerId);

        assertThat(response.status()).isEqualTo(SellerStatus.SUSPENDED);
    }

    // -------------------------------------------------------------------------
    // Test 6: Tier recalculation on stats update
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("6 — updateSellerStats recalculates tier correctly (50 orders → SILVER)")
    void updateSellerStats_50Orders_tierBecomessilver() {
        when(sellerRepository.findById(sellerId)).thenReturn(Optional.of(activeSeller));
        when(sellerRepository.save(any(Seller.class))).thenAnswer(inv -> inv.getArgument(0));

        SellerResponse response = sellerService.updateSellerStats(
                sellerId, 50, new BigDecimal("5000.00"), new BigDecimal("2.00"));

        assertThat(response.tier()).isEqualTo(SellerTier.SILVER);
        assertThat(response.totalOrders()).isEqualTo(50);
    }

    // -------------------------------------------------------------------------
    // Test 7: Tier recalculation — 1000 orders → PLATINUM
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("7 — updateSellerStats with 1000 cumulative orders sets PLATINUM tier")
    void updateSellerStats_1000Orders_tierBecomesPlatinum() {
        activeSeller.setTotalOrders(950);
        when(sellerRepository.findById(sellerId)).thenReturn(Optional.of(activeSeller));
        when(sellerRepository.save(any(Seller.class))).thenAnswer(inv -> inv.getArgument(0));

        SellerResponse response = sellerService.updateSellerStats(
                sellerId, 50, new BigDecimal("100000.00"), new BigDecimal("1.00"));

        assertThat(response.tier()).isEqualTo(SellerTier.PLATINUM);
    }

    // -------------------------------------------------------------------------
    // Test 8: Create listing for ACTIVE seller
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("8 — createListing for ACTIVE seller persists listing in PENDING status")
    void createListing_activeSeller_persistsListing() {
        ListingRequest request = new ListingRequest(
                "PROD-123", "SKU-001", "SELLER-SKU-001", new BigDecimal("29.99"), 100);

        SellerProduct savedProduct = SellerProduct.builder()
                .id(UUID.randomUUID())
                .sellerId(sellerId)
                .productId("PROD-123")
                .sku("SKU-001")
                .sellerSku("SELLER-SKU-001")
                .listingPrice(new BigDecimal("29.99"))
                .status("PENDING")
                .stockQuantity(100)
                .build();

        when(sellerRepository.findById(sellerId)).thenReturn(Optional.of(activeSeller));
        when(sellerProductRepository.findBySellerIdAndSellerSku(sellerId, "SELLER-SKU-001"))
                .thenReturn(Optional.empty());
        when(sellerProductRepository.save(any(SellerProduct.class))).thenReturn(savedProduct);
        when(sellerRepository.save(any(Seller.class))).thenReturn(activeSeller);

        SellerProduct result = sellerService.createListing(sellerId, request);

        assertThat(result.getStatus()).isEqualTo("PENDING");
        assertThat(result.getListingPrice()).isEqualByComparingTo("29.99");
        verify(sellerProductRepository).save(any(SellerProduct.class));
    }

    // -------------------------------------------------------------------------
    // Test 9: Deactivate listing sets status to INACTIVE
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("9 — deactivateListing sets listing status to INACTIVE")
    void deactivateListing_activeListing_setsInactiveStatus() {
        UUID listingId = UUID.randomUUID();
        SellerProduct product = SellerProduct.builder()
                .id(listingId)
                .sellerId(sellerId)
                .productId("PROD-X")
                .sku("SKU-X")
                .sellerSku("SELLER-X")
                .listingPrice(new BigDecimal("19.99"))
                .status("ACTIVE")
                .stockQuantity(50)
                .build();

        when(sellerProductRepository.findById(listingId)).thenReturn(Optional.of(product));
        when(sellerProductRepository.save(any(SellerProduct.class)))
                .thenAnswer(inv -> inv.getArgument(0));

        SellerProduct result = sellerService.deactivateListing(listingId);

        assertThat(result.getStatus()).isEqualTo("INACTIVE");
        verify(sellerProductRepository).save(product);
    }

    // -------------------------------------------------------------------------
    // Test 10: listSellers with status filter returns filtered results
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("10 — listSellers with ACTIVE status filter delegates to repository and maps to responses")
    void listSellers_activeStatusFilter_returnsFilteredSellers() {
        List<Seller> activeSellers = List.of(activeSeller);
        when(sellerRepository.findByStatus(SellerStatus.ACTIVE)).thenReturn(activeSellers);

        List<SellerResponse> result = sellerService.listSellers(SellerStatus.ACTIVE, null);

        assertThat(result).hasSize(1);
        assertThat(result.get(0).status()).isEqualTo(SellerStatus.ACTIVE);
        verify(sellerRepository).findByStatus(SellerStatus.ACTIVE);
        verify(sellerRepository, never()).findAll();
    }
}
