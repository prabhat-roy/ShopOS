package com.shopos.adservice.repository;

import com.shopos.adservice.domain.AdImpression;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.UUID;

@Repository
public interface AdImpressionRepository extends JpaRepository<AdImpression, UUID> {

    Page<AdImpression> findByCampaignId(UUID campaignId, Pageable pageable);

    long countByCampaignId(UUID campaignId);

    long countByCampaignIdAndClickedTrue(UUID campaignId);
}
