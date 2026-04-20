package com.shopos.pricelistservice.service;

import com.shopos.pricelistservice.domain.PriceList;
import com.shopos.pricelistservice.domain.PriceListEntry;
import com.shopos.pricelistservice.dto.CreatePriceListRequest;
import com.shopos.pricelistservice.dto.SetEntryRequest;
import com.shopos.pricelistservice.dto.UpdatePriceListRequest;
import com.shopos.pricelistservice.repository.PriceListEntryRepository;
import com.shopos.pricelistservice.repository.PriceListRepository;
import jakarta.persistence.EntityNotFoundException;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class PriceListService {

    private final PriceListRepository priceListRepository;
    private final PriceListEntryRepository priceListEntryRepository;

    // ── Price List CRUD ───────────────────────────────────────────────────────

    @Transactional
    public PriceList createList(CreatePriceListRequest req) {
        PriceList list = PriceList.builder()
                .name(req.name())
                .code(req.code().toUpperCase())
                .currency(req.currency())
                .description(req.description() != null ? req.description() : "")
                .active(true)
                .build();
        return priceListRepository.save(list);
    }

    @Transactional(readOnly = true)
    public PriceList getList(UUID id) {
        return priceListRepository.findById(id)
                .orElseThrow(() -> new EntityNotFoundException("PriceList not found with id=" + id));
    }

    @Transactional(readOnly = true)
    public List<PriceList> listLists() {
        return priceListRepository.findByActiveTrue();
    }

    @Transactional
    public PriceList updateList(UUID id, UpdatePriceListRequest req) {
        PriceList list = priceListRepository.findById(id)
                .orElseThrow(() -> new EntityNotFoundException("PriceList not found with id=" + id));

        if (req.name() != null && !req.name().isBlank()) {
            list.setName(req.name());
        }
        if (req.description() != null) {
            list.setDescription(req.description());
        }
        if (req.active() != null) {
            list.setActive(req.active());
        }

        return priceListRepository.save(list);
    }

    @Transactional
    public void deleteList(UUID id) {
        PriceList list = priceListRepository.findById(id)
                .orElseThrow(() -> new EntityNotFoundException("PriceList not found with id=" + id));
        list.setActive(false);
        priceListRepository.save(list);
    }

    // ── Price List Entry operations ──────────────────────────────────────────

    /**
     * Creates or updates the price for a product in the given price list.
     */
    @Transactional
    public PriceListEntry setEntry(UUID listId, SetEntryRequest req) {
        // Verify the list exists
        if (!priceListRepository.existsById(listId)) {
            throw new EntityNotFoundException("PriceList not found with id=" + listId);
        }

        PriceListEntry entry = priceListEntryRepository
                .findByPriceListIdAndProductId(listId, req.productId())
                .orElse(PriceListEntry.builder()
                        .priceListId(listId)
                        .productId(req.productId())
                        .build());

        entry.setPrice(req.price());
        return priceListEntryRepository.save(entry);
    }

    @Transactional(readOnly = true)
    public PriceListEntry getEntry(UUID listId, String productId) {
        return priceListEntryRepository.findByPriceListIdAndProductId(listId, productId)
                .orElseThrow(() -> new EntityNotFoundException(
                        "No price entry found for productId=" + productId
                                + " in priceListId=" + listId));
    }

    @Transactional
    public void removeEntry(UUID listId, String productId) {
        PriceListEntry entry = priceListEntryRepository
                .findByPriceListIdAndProductId(listId, productId)
                .orElseThrow(() -> new EntityNotFoundException(
                        "No price entry found for productId=" + productId
                                + " in priceListId=" + listId));
        priceListEntryRepository.delete(entry);
    }

    @Transactional(readOnly = true)
    public List<PriceListEntry> listEntries(UUID listId) {
        if (!priceListRepository.existsById(listId)) {
            throw new EntityNotFoundException("PriceList not found with id=" + listId);
        }
        return priceListEntryRepository.findByPriceListId(listId);
    }

    /**
     * Looks up the price of a product in a named price list by code.
     * Used for B2B pricing lookups: GET /price-lists/lookup?code=WHOLESALE&productId=prod-001
     */
    @Transactional(readOnly = true)
    public PriceListEntry getProductPrice(String listCode, String productId) {
        PriceList list = priceListRepository.findByCode(listCode.toUpperCase())
                .orElseThrow(() -> new EntityNotFoundException(
                        "PriceList not found with code=" + listCode));

        return priceListEntryRepository
                .findByPriceListIdAndProductId(list.getId(), productId)
                .orElseThrow(() -> new EntityNotFoundException(
                        "No price entry found for productId=" + productId
                                + " in price list '" + listCode + "'"));
    }
}
