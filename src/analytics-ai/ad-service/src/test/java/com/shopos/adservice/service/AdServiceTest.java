package com.shopos.adservice.service;

import com.shopos.adservice.domain.AdCampaign;
import com.shopos.adservice.domain.AdImpression;
import com.shopos.adservice.domain.AdStatus;
import com.shopos.adservice.domain.AdType;
import com.shopos.adservice.dto.*;
import com.shopos.adservice.exception.ResourceNotFoundException;
import com.shopos.adservice.repository.AdCampaignRepository;
import com.shopos.adservice.repository.AdImpressionRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AdServiceTest {

    @Mock
    private AdCampaignRepository campaignRepository;

    @Mock
    private AdImpressionRepository impressionRepository;

    @InjectMocks
    private AdService adService;

    private UUID campaignId;
    private AdCampaign activeCampaign;

    @BeforeEach
    void setUp() {
        campaignId = UUID.randomUUID();
        activeCampaign = AdCampaign.builder()
                .id(campaignId)
                .name("Test Campaign")
                .advertiserId(UUID.randomUUID())
                .status(AdStatus.ACTIVE)
                .adType(AdType.BANNER)
                .budget(new BigDecimal("1000.00"))
                .spent(BigDecimal.ZERO)
                .impressions(0L)
                .clicks(0L)
                .startDate(LocalDate.now().minusDays(1))
                .endDate(LocalDate.now().plusDays(30))
                .targetUrl("https://example.com")
                .bidAmount(new BigDecimal("2.50"))
                .build();
    }

    @Test
    void createCampaign_shouldReturnCreatedCampaignWithDraftStatus() {
        CreateCampaignRequest req = new CreateCampaignRequest();
        req.setName("New Banner");
        req.setAdvertiserId(UUID.randomUUID());
        req.setAdType(AdType.BANNER);
        req.setBudget(new BigDecimal("500.00"));
        req.setStartDate(LocalDate.now());
        req.setEndDate(LocalDate.now().plusDays(7));
        req.setTargetUrl("https://shop.example.com");
        req.setBidAmount(new BigDecimal("1.00"));

        AdCampaign saved = AdCampaign.builder()
                .id(UUID.randomUUID())
                .name(req.getName())
                .advertiserId(req.getAdvertiserId())
                .status(AdStatus.DRAFT)
                .adType(AdType.BANNER)
                .budget(req.getBudget())
                .spent(BigDecimal.ZERO)
                .impressions(0L)
                .clicks(0L)
                .startDate(req.getStartDate())
                .endDate(req.getEndDate())
                .targetUrl(req.getTargetUrl())
                .bidAmount(req.getBidAmount())
                .build();

        when(campaignRepository.save(any(AdCampaign.class))).thenReturn(saved);

        CampaignResponse response = adService.createCampaign(req);

        assertThat(response.getStatus()).isEqualTo(AdStatus.DRAFT);
        assertThat(response.getName()).isEqualTo("New Banner");
        verify(campaignRepository, times(1)).save(any());
    }

    @Test
    void getCampaign_shouldReturnCampaign_whenExists() {
        when(campaignRepository.findById(campaignId)).thenReturn(Optional.of(activeCampaign));

        CampaignResponse response = adService.getCampaign(campaignId);

        assertThat(response.getId()).isEqualTo(campaignId);
        assertThat(response.getName()).isEqualTo("Test Campaign");
    }

    @Test
    void getCampaign_shouldThrow_whenNotFound() {
        when(campaignRepository.findById(any())).thenReturn(Optional.empty());

        assertThatThrownBy(() -> adService.getCampaign(UUID.randomUUID()))
                .isInstanceOf(ResourceNotFoundException.class);
    }

    @Test
    void activateCampaign_shouldSetStatusToActive() {
        AdCampaign draftCampaign = AdCampaign.builder()
                .id(campaignId)
                .status(AdStatus.DRAFT)
                .budget(BigDecimal.TEN)
                .spent(BigDecimal.ZERO)
                .impressions(0L)
                .clicks(0L)
                .name("Draft")
                .adType(AdType.NATIVE)
                .startDate(LocalDate.now())
                .endDate(LocalDate.now().plusDays(10))
                .targetUrl("https://x.com")
                .build();

        when(campaignRepository.findById(campaignId)).thenReturn(Optional.of(draftCampaign));
        when(campaignRepository.save(any())).thenReturn(draftCampaign);

        CampaignResponse response = adService.activateCampaign(campaignId);

        assertThat(draftCampaign.getStatus()).isEqualTo(AdStatus.ACTIVE);
    }

    @Test
    void pauseCampaign_shouldSetStatusToPaused() {
        when(campaignRepository.findById(campaignId)).thenReturn(Optional.of(activeCampaign));
        when(campaignRepository.save(any())).thenReturn(activeCampaign);

        adService.pauseCampaign(campaignId);

        assertThat(activeCampaign.getStatus()).isEqualTo(AdStatus.PAUSED);
        verify(campaignRepository).save(activeCampaign);
    }

    @Test
    void pauseCampaign_shouldThrow_whenNotActive() {
        activeCampaign.setStatus(AdStatus.DRAFT);
        when(campaignRepository.findById(campaignId)).thenReturn(Optional.of(activeCampaign));

        assertThatThrownBy(() -> adService.pauseCampaign(campaignId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("ACTIVE");
    }

    @Test
    void cancelCampaign_shouldSetStatusToCancelled() {
        when(campaignRepository.findById(campaignId)).thenReturn(Optional.of(activeCampaign));
        when(campaignRepository.save(any())).thenReturn(activeCampaign);

        adService.cancelCampaign(campaignId);

        assertThat(activeCampaign.getStatus()).isEqualTo(AdStatus.CANCELLED);
    }

    @Test
    void serveAd_shouldReturnCampaign_whenMatchFound() {
        when(campaignRepository.findActiveAdsByType(any(LocalDate.class), eq(AdType.BANNER)))
                .thenReturn(List.of(activeCampaign));
        when(impressionRepository.save(any())).thenReturn(AdImpression.builder()
                .id(UUID.randomUUID())
                .campaignId(campaignId)
                .sessionId("sess-1")
                .placementId("home-hero")
                .build());
        when(campaignRepository.save(any())).thenReturn(activeCampaign);

        ServeAdRequest req = new ServeAdRequest("user-1", "sess-1", "home-hero", List.of("electronics"), AdType.BANNER);
        Optional<CampaignResponse> result = adService.serveAd(req);

        assertThat(result).isPresent();
        assertThat(result.get().getId()).isEqualTo(campaignId);
        verify(impressionRepository).save(any());
    }

    @Test
    void serveAd_shouldReturnEmpty_whenNoCandidates() {
        when(campaignRepository.findActiveAdsByType(any(LocalDate.class), any()))
                .thenReturn(List.of());

        ServeAdRequest req = new ServeAdRequest("user-1", "sess-1", "sidebar", List.of(), AdType.VIDEO);
        Optional<CampaignResponse> result = adService.serveAd(req);

        assertThat(result).isEmpty();
        verify(impressionRepository, never()).save(any());
    }

    @Test
    void recordClick_shouldMarkImpressionClickedAndIncrementCampaign() {
        UUID impressionId = UUID.randomUUID();
        AdImpression impression = AdImpression.builder()
                .id(impressionId)
                .campaignId(campaignId)
                .sessionId("sess-1")
                .placementId("home")
                .clicked(false)
                .build();

        when(impressionRepository.findById(impressionId)).thenReturn(Optional.of(impression));
        when(impressionRepository.save(any())).thenReturn(impression);
        when(campaignRepository.findById(campaignId)).thenReturn(Optional.of(activeCampaign));
        when(campaignRepository.save(any())).thenReturn(activeCampaign);

        adService.recordClick(impressionId);

        assertThat(impression.isClicked()).isTrue();
        assertThat(impression.getClickedAt()).isNotNull();
        assertThat(activeCampaign.getClicks()).isEqualTo(1L);
    }

    @Test
    void getCampaignStats_shouldComputeCtrCorrectly() {
        when(campaignRepository.findById(campaignId)).thenReturn(Optional.of(activeCampaign));
        when(impressionRepository.countByCampaignId(campaignId)).thenReturn(200L);
        when(impressionRepository.countByCampaignIdAndClickedTrue(campaignId)).thenReturn(10L);

        CampaignStatsResponse stats = adService.getCampaignStats(campaignId);

        assertThat(stats.getImpressions()).isEqualTo(200L);
        assertThat(stats.getClicks()).isEqualTo(10L);
        assertThat(stats.getCtr()).isEqualTo(0.05);
    }
}
