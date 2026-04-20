package com.shopos.adservice.repository;

import com.shopos.adservice.domain.AdCampaign;
import com.shopos.adservice.domain.AdStatus;
import com.shopos.adservice.domain.AdType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.UUID;

@Repository
public interface AdCampaignRepository extends JpaRepository<AdCampaign, UUID> {

    List<AdCampaign> findByAdvertiserId(UUID advertiserId);

    List<AdCampaign> findByStatus(AdStatus status);

    @Query("SELECT c FROM AdCampaign c WHERE c.status = 'ACTIVE' " +
           "AND c.startDate <= :today AND c.endDate >= :today")
    List<AdCampaign> findActiveAds(@Param("today") LocalDate today);

    @Query("SELECT c FROM AdCampaign c WHERE c.status = 'ACTIVE' " +
           "AND c.startDate <= :today AND c.endDate >= :today " +
           "AND c.adType = :adType")
    List<AdCampaign> findActiveAdsByType(@Param("today") LocalDate today,
                                         @Param("adType") AdType adType);

    @Query("SELECT c FROM AdCampaign c WHERE c.status = 'ACTIVE' " +
           "AND c.startDate <= :today AND c.endDate >= :today " +
           "AND (c.targetCategories IS NULL OR c.targetCategories LIKE %:categoryId%)")
    List<AdCampaign> findByCategoryTarget(@Param("today") LocalDate today,
                                          @Param("categoryId") String categoryId);
}
