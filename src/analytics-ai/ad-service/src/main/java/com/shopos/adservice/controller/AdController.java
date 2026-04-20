package com.shopos.adservice.controller;

import com.shopos.adservice.domain.AdStatus;
import com.shopos.adservice.dto.*;
import com.shopos.adservice.service.AdService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequiredArgsConstructor
public class AdController {

    private final AdService adService;

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> healthz() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    @PostMapping("/campaigns")
    public ResponseEntity<CampaignResponse> createCampaign(
            @Valid @RequestBody CreateCampaignRequest request) {
        return ResponseEntity.status(HttpStatus.CREATED).body(adService.createCampaign(request));
    }

    @GetMapping("/campaigns/{id}")
    public ResponseEntity<CampaignResponse> getCampaign(@PathVariable UUID id) {
        return ResponseEntity.ok(adService.getCampaign(id));
    }

    @GetMapping("/campaigns")
    public ResponseEntity<List<CampaignResponse>> listCampaigns(
            @RequestParam(required = false) UUID advertiserId,
            @RequestParam(required = false) AdStatus status) {
        return ResponseEntity.ok(adService.listCampaigns(advertiserId, status));
    }

    @PostMapping("/campaigns/{id}/activate")
    public ResponseEntity<Void> activateCampaign(@PathVariable UUID id) {
        adService.activateCampaign(id);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/campaigns/{id}/pause")
    public ResponseEntity<Void> pauseCampaign(@PathVariable UUID id) {
        adService.pauseCampaign(id);
        return ResponseEntity.noContent().build();
    }

    @DeleteMapping("/campaigns/{id}")
    public ResponseEntity<Void> cancelCampaign(@PathVariable UUID id) {
        adService.cancelCampaign(id);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/ads/serve")
    public ResponseEntity<?> serveAd(@Valid @RequestBody ServeAdRequest request) {
        return adService.serveAd(request)
                .<ResponseEntity<?>>map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.noContent().build());
    }

    @PostMapping("/ads/click/{impressionId}")
    public ResponseEntity<Void> recordClick(@PathVariable UUID impressionId) {
        adService.recordClick(impressionId);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/campaigns/{id}/stats")
    public ResponseEntity<CampaignStatsResponse> getCampaignStats(@PathVariable UUID id) {
        return ResponseEntity.ok(adService.getCampaignStats(id));
    }
}
