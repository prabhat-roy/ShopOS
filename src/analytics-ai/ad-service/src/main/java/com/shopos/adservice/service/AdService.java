package com.shopos.adservice.service;

import com.shopos.adservice.domain.AdCampaign;
import com.shopos.adservice.domain.AdImpression;
import com.shopos.adservice.domain.AdStatus;
import com.shopos.adservice.dto.*;
import com.shopos.adservice.exception.ResourceNotFoundException;
import com.shopos.adservice.repository.AdCampaignRepository;
import com.shopos.adservice.repository.AdImpressionRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class AdService {

    private final AdCampaignRepository campaignRepository;
    private final AdImpressionRepository impressionRepository;

    @Transactional
    public CampaignResponse createCampaign(CreateCampaignRequest request) {
        AdCampaign campaign = AdCampaign.builder()
                .name(request.getName())
                .advertiserId(request.getAdvertiserId())
                .status(AdStatus.DRAFT)
                .adType(request.getAdType())
                .targetCategories(request.getTargetCategories())
                .targetAudience(request.getTargetAudience())
                .budget(request.getBudget())
                .startDate(request.getStartDate())
                .endDate(request.getEndDate())
                .imageUrl(request.getImageUrl())
                .targetUrl(request.getTargetUrl())
                .bidAmount(request.getBidAmount())
                .build();
        AdCampaign saved = campaignRepository.save(campaign);
        log.info("Created campaign id={} name={}", saved.getId(), saved.getName());
        return CampaignResponse.from(saved);
    }

    @Transactional(readOnly = true)
    public CampaignResponse getCampaign(UUID id) {
        return CampaignResponse.from(findCampaignOrThrow(id));
    }

    @Transactional(readOnly = true)
    public List<CampaignResponse> listCampaigns(UUID advertiserId, AdStatus status) {
        List<AdCampaign> results;
        if (advertiserId != null) {
            results = campaignRepository.findByAdvertiserId(advertiserId);
        } else if (status != null) {
            results = campaignRepository.findByStatus(status);
        } else {
            results = campaignRepository.findAll();
        }
        return results.stream().map(CampaignResponse::from).collect(Collectors.toList());
    }

    @Transactional
    public CampaignResponse activateCampaign(UUID id) {
        AdCampaign campaign = findCampaignOrThrow(id);
        if (campaign.getStatus() == AdStatus.CANCELLED) {
            throw new IllegalStateException("Cannot activate a cancelled campaign");
        }
        campaign.setStatus(AdStatus.ACTIVE);
        return CampaignResponse.from(campaignRepository.save(campaign));
    }

    @Transactional
    public CampaignResponse pauseCampaign(UUID id) {
        AdCampaign campaign = findCampaignOrThrow(id);
        if (campaign.getStatus() != AdStatus.ACTIVE) {
            throw new IllegalStateException("Only ACTIVE campaigns can be paused");
        }
        campaign.setStatus(AdStatus.PAUSED);
        return CampaignResponse.from(campaignRepository.save(campaign));
    }

    @Transactional
    public void cancelCampaign(UUID id) {
        AdCampaign campaign = findCampaignOrThrow(id);
        if (campaign.getStatus() == AdStatus.COMPLETED) {
            throw new IllegalStateException("Cannot cancel a completed campaign");
        }
        campaign.setStatus(AdStatus.CANCELLED);
        campaignRepository.save(campaign);
        log.info("Cancelled campaign id={}", id);
    }

    @Transactional
    public Optional<CampaignResponse> serveAd(ServeAdRequest request) {
        LocalDate today = LocalDate.now();

        List<AdCampaign> candidates = campaignRepository.findActiveAdsByType(today, request.adType());

        // Filter by category overlap if categories are provided
        if (request.categories() != null && !request.categories().isEmpty()) {
            candidates = candidates.stream()
                    .filter(c -> {
                        if (c.getTargetCategories() == null || c.getTargetCategories().isBlank()) return true;
                        List<String> targets = List.of(c.getTargetCategories().split(","));
                        return request.categories().stream().anyMatch(targets::contains);
                    })
                    .collect(Collectors.toList());
        }

        if (candidates.isEmpty()) {
            return Optional.empty();
        }

        // Pick the campaign with the highest bid amount (simple auction)
        AdCampaign winner = candidates.stream()
                .filter(c -> c.getBidAmount() != null)
                .max((a, b) -> a.getBidAmount().compareTo(b.getBidAmount()))
                .orElse(candidates.get(0));

        // Record impression
        AdImpression impression = AdImpression.builder()
                .campaignId(winner.getId())
                .userId(request.userId())
                .sessionId(request.sessionId())
                .placementId(request.placementId())
                .build();
        impressionRepository.save(impression);

        // Increment impression counter and deduct cost
        winner.setImpressions(winner.getImpressions() + 1);
        BigDecimal cpm = winner.getBidAmount() != null ? winner.getBidAmount().divide(BigDecimal.valueOf(1000), 4, RoundingMode.HALF_UP) : BigDecimal.ZERO;
        winner.setSpent(winner.getSpent().add(cpm));
        campaignRepository.save(winner);

        log.debug("Served ad campaign={} to session={}", winner.getId(), request.sessionId());
        return Optional.of(CampaignResponse.from(winner));
    }

    @Transactional
    public void recordClick(UUID impressionId) {
        AdImpression impression = impressionRepository.findById(impressionId)
                .orElseThrow(() -> new ResourceNotFoundException("Impression not found: " + impressionId));

        if (impression.isClicked()) {
            log.warn("Impression {} already clicked, ignoring duplicate", impressionId);
            return;
        }

        impression.setClicked(true);
        impression.setClickedAt(LocalDateTime.now());
        impressionRepository.save(impression);

        AdCampaign campaign = findCampaignOrThrow(impression.getCampaignId());
        campaign.setClicks(campaign.getClicks() + 1);
        campaignRepository.save(campaign);
    }

    @Transactional(readOnly = true)
    public CampaignStatsResponse getCampaignStats(UUID id) {
        AdCampaign campaign = findCampaignOrThrow(id);
        long totalImpressions = impressionRepository.countByCampaignId(id);
        long totalClicks      = impressionRepository.countByCampaignIdAndClickedTrue(id);

        double ctr = totalImpressions == 0 ? 0.0
                : BigDecimal.valueOf((double) totalClicks / totalImpressions)
                            .setScale(4, RoundingMode.HALF_UP)
                            .doubleValue();

        double budgetUtil = campaign.getBudget().compareTo(BigDecimal.ZERO) == 0 ? 0.0
                : campaign.getSpent()
                          .divide(campaign.getBudget(), 4, RoundingMode.HALF_UP)
                          .multiply(BigDecimal.valueOf(100))
                          .doubleValue();

        return CampaignStatsResponse.builder()
                .campaignId(id)
                .impressions(totalImpressions)
                .clicks(totalClicks)
                .ctr(ctr)
                .spent(campaign.getSpent())
                .budget(campaign.getBudget())
                .budgetUtilizationPct(budgetUtil)
                .build();
    }

    private AdCampaign findCampaignOrThrow(UUID id) {
        return campaignRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("Campaign not found: " + id));
    }
}
