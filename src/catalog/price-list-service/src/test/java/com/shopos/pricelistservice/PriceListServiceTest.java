package com.shopos.pricelistservice;

import com.shopos.pricelistservice.domain.PriceList;
import com.shopos.pricelistservice.domain.PriceListEntry;
import com.shopos.pricelistservice.dto.CreatePriceListRequest;
import com.shopos.pricelistservice.dto.SetEntryRequest;
import com.shopos.pricelistservice.repository.PriceListEntryRepository;
import com.shopos.pricelistservice.repository.PriceListRepository;
import com.shopos.pricelistservice.service.PriceListService;
import jakarta.persistence.EntityNotFoundException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class PriceListServiceTest {

    @Mock
    private PriceListRepository priceListRepository;

    @Mock
    private PriceListEntryRepository priceListEntryRepository;

    @InjectMocks
    private PriceListService priceListService;

    private PriceList wholesaleList;

    @BeforeEach
    void setUp() {
        wholesaleList = PriceList.builder()
                .id(UUID.randomUUID())
                .name("Wholesale")
                .code("WHOLESALE")
                .currency("USD")
                .description("B2B wholesale pricing")
                .active(true)
                .build();
    }

    // ── createList ──────────────────────────────────────────────────────────

    @Test
    @DisplayName("createList persists and returns the new price list")
    void createList_validRequest_returnsSavedList() {
        CreatePriceListRequest req = new CreatePriceListRequest(
                "Wholesale", "WHOLESALE", "USD", "B2B wholesale pricing");

        when(priceListRepository.save(any(PriceList.class))).thenReturn(wholesaleList);

        PriceList result = priceListService.createList(req);

        assertThat(result).isNotNull();
        assertThat(result.getCode()).isEqualTo("WHOLESALE");
        assertThat(result.getName()).isEqualTo("Wholesale");
        assertThat(result.getCurrency()).isEqualTo("USD");
        assertThat(result.isActive()).isTrue();

        verify(priceListRepository, times(1)).save(any(PriceList.class));
    }

    @Test
    @DisplayName("createList uppercases the code before saving")
    void createList_lowercaseCode_isSavedAsUppercase() {
        CreatePriceListRequest req = new CreatePriceListRequest(
                "VIP", "vip", "EUR", "VIP customer pricing");

        PriceList vipList = PriceList.builder()
                .id(UUID.randomUUID())
                .name("VIP")
                .code("VIP")
                .currency("EUR")
                .active(true)
                .build();

        when(priceListRepository.save(any(PriceList.class))).thenReturn(vipList);

        PriceList result = priceListService.createList(req);

        assertThat(result.getCode()).isEqualTo("VIP");
    }

    // ── getList ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("getList returns the price list when it exists")
    void getList_existingId_returnsPriceList() {
        UUID id = wholesaleList.getId();
        when(priceListRepository.findById(id)).thenReturn(Optional.of(wholesaleList));

        PriceList result = priceListService.getList(id);

        assertThat(result.getId()).isEqualTo(id);
        assertThat(result.getCode()).isEqualTo("WHOLESALE");
    }

    @Test
    @DisplayName("getList throws EntityNotFoundException for unknown ID")
    void getList_unknownId_throwsEntityNotFoundException() {
        UUID unknown = UUID.randomUUID();
        when(priceListRepository.findById(unknown)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> priceListService.getList(unknown))
                .isInstanceOf(EntityNotFoundException.class)
                .hasMessageContaining(unknown.toString());
    }

    // ── getProductPrice ──────────────────────────────────────────────────────

    @Test
    @DisplayName("getProductPrice returns correct entry price for product in named list")
    void getProductPrice_validCodeAndProduct_returnsEntry() {
        String productId = "prod-001";
        BigDecimal expectedPrice = new BigDecimal("45.00");

        PriceListEntry entry = PriceListEntry.builder()
                .id(UUID.randomUUID())
                .priceListId(wholesaleList.getId())
                .productId(productId)
                .price(expectedPrice)
                .build();

        when(priceListRepository.findByCode("WHOLESALE")).thenReturn(Optional.of(wholesaleList));
        when(priceListEntryRepository.findByPriceListIdAndProductId(wholesaleList.getId(), productId))
                .thenReturn(Optional.of(entry));

        PriceListEntry result = priceListService.getProductPrice("WHOLESALE", productId);

        assertThat(result.getPrice()).isEqualByComparingTo("45.00");
        assertThat(result.getProductId()).isEqualTo(productId);
    }

    @Test
    @DisplayName("getProductPrice throws EntityNotFoundException for unknown list code")
    void getProductPrice_unknownCode_throwsEntityNotFoundException() {
        when(priceListRepository.findByCode("NONEXISTENT")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> priceListService.getProductPrice("NONEXISTENT", "prod-001"))
                .isInstanceOf(EntityNotFoundException.class)
                .hasMessageContaining("NONEXISTENT");
    }

    @Test
    @DisplayName("getProductPrice throws EntityNotFoundException when product has no entry in the list")
    void getProductPrice_productNotInList_throwsEntityNotFoundException() {
        when(priceListRepository.findByCode("WHOLESALE")).thenReturn(Optional.of(wholesaleList));
        when(priceListEntryRepository.findByPriceListIdAndProductId(wholesaleList.getId(), "prod-999"))
                .thenReturn(Optional.empty());

        assertThatThrownBy(() -> priceListService.getProductPrice("WHOLESALE", "prod-999"))
                .isInstanceOf(EntityNotFoundException.class)
                .hasMessageContaining("prod-999");
    }

    // ── setEntry ────────────────────────────────────────────────────────────

    @Test
    @DisplayName("setEntry creates a new entry when none exists for the product")
    void setEntry_newProduct_createsEntry() {
        UUID listId = wholesaleList.getId();
        SetEntryRequest req = new SetEntryRequest("prod-100", new BigDecimal("55.00"));

        PriceListEntry savedEntry = PriceListEntry.builder()
                .id(UUID.randomUUID())
                .priceListId(listId)
                .productId("prod-100")
                .price(new BigDecimal("55.00"))
                .build();

        when(priceListRepository.existsById(listId)).thenReturn(true);
        when(priceListEntryRepository.findByPriceListIdAndProductId(listId, "prod-100"))
                .thenReturn(Optional.empty());
        when(priceListEntryRepository.save(any(PriceListEntry.class))).thenReturn(savedEntry);

        PriceListEntry result = priceListService.setEntry(listId, req);

        assertThat(result.getPrice()).isEqualByComparingTo("55.00");
        assertThat(result.getProductId()).isEqualTo("prod-100");
        verify(priceListEntryRepository, times(1)).save(any(PriceListEntry.class));
    }

    @Test
    @DisplayName("setEntry updates price when entry already exists for the product")
    void setEntry_existingProduct_updatesPrice() {
        UUID listId = wholesaleList.getId();
        SetEntryRequest req = new SetEntryRequest("prod-001", new BigDecimal("40.00"));

        PriceListEntry existing = PriceListEntry.builder()
                .id(UUID.randomUUID())
                .priceListId(listId)
                .productId("prod-001")
                .price(new BigDecimal("45.00"))
                .build();

        PriceListEntry updated = PriceListEntry.builder()
                .id(existing.getId())
                .priceListId(listId)
                .productId("prod-001")
                .price(new BigDecimal("40.00"))
                .build();

        when(priceListRepository.existsById(listId)).thenReturn(true);
        when(priceListEntryRepository.findByPriceListIdAndProductId(listId, "prod-001"))
                .thenReturn(Optional.of(existing));
        when(priceListEntryRepository.save(any(PriceListEntry.class))).thenReturn(updated);

        PriceListEntry result = priceListService.setEntry(listId, req);

        assertThat(result.getPrice()).isEqualByComparingTo("40.00");
    }
}
