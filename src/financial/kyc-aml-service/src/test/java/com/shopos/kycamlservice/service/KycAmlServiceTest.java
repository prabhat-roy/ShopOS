package com.shopos.kycamlservice.service;

import com.shopos.kycamlservice.domain.AmlCheck;
import com.shopos.kycamlservice.domain.KycRecord;
import com.shopos.kycamlservice.domain.KycStatus;
import com.shopos.kycamlservice.domain.RiskLevel;
import com.shopos.kycamlservice.dto.*;
import com.shopos.kycamlservice.exception.ConflictException;
import com.shopos.kycamlservice.exception.InvalidStateTransitionException;
import com.shopos.kycamlservice.exception.ResourceNotFoundException;
import com.shopos.kycamlservice.repository.AmlCheckRepository;
import com.shopos.kycamlservice.repository.KycRecordRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class KycAmlServiceTest {

    @Mock
    private KycRecordRepository kycRecordRepository;

    @Mock
    private AmlCheckRepository amlCheckRepository;

    @InjectMocks
    private KycAmlService kycAmlService;

    private UUID customerId;
    private UUID kycId;
    private UUID amlId;

    @BeforeEach
    void setUp() {
        customerId = UUID.randomUUID();
        kycId = UUID.randomUUID();
        amlId = UUID.randomUUID();
    }

    // -------------------------------------------------------------------------
    // Test 1: createKycRecord — happy path
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("createKycRecord saves new record with PENDING status")
    void createKycRecord_savesNewRecordWithPendingStatus() {
        CreateKycRequest request = new CreateKycRequest(
            customerId, "John", "Doe",
            LocalDate.of(1985, 6, 15), "US",
            "PASSPORT", "X12345678",
            LocalDate.now().plusYears(5), null
        );

        KycRecord savedRecord = buildKycRecord(kycId, customerId, KycStatus.PENDING, "US");

        when(kycRecordRepository.existsByCustomerId(customerId)).thenReturn(false);
        when(kycRecordRepository.save(any(KycRecord.class))).thenReturn(savedRecord);

        KycResponse response = kycAmlService.createKycRecord(request);

        assertThat(response).isNotNull();
        assertThat(response.customerId()).isEqualTo(customerId);
        assertThat(response.status()).isEqualTo(KycStatus.PENDING);
        assertThat(response.firstName()).isEqualTo("John");
        verify(kycRecordRepository).save(any(KycRecord.class));
    }

    // -------------------------------------------------------------------------
    // Test 2: createKycRecord — duplicate customerId throws ConflictException
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("createKycRecord throws ConflictException when customerId already exists")
    void createKycRecord_throwsConflict_whenCustomerIdAlreadyExists() {
        CreateKycRequest request = new CreateKycRequest(
            customerId, "Jane", "Smith",
            LocalDate.of(1990, 3, 20), "GB",
            "NATIONAL_ID", "N987654321",
            LocalDate.now().plusYears(3), null
        );

        when(kycRecordRepository.existsByCustomerId(customerId)).thenReturn(true);

        assertThatThrownBy(() -> kycAmlService.createKycRecord(request))
            .isInstanceOf(ConflictException.class)
            .hasMessageContaining(customerId.toString());

        verify(kycRecordRepository, never()).save(any());
    }

    // -------------------------------------------------------------------------
    // Test 3: getByCustomerId — returns response when record exists
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("getByCustomerId returns KycResponse for existing customer")
    void getByCustomerId_returnsResponse_whenRecordExists() {
        KycRecord record = buildKycRecord(kycId, customerId, KycStatus.VERIFIED, "DE");

        when(kycRecordRepository.findByCustomerId(customerId)).thenReturn(Optional.of(record));

        KycResponse response = kycAmlService.getByCustomerId(customerId);

        assertThat(response.id()).isEqualTo(kycId);
        assertThat(response.customerId()).isEqualTo(customerId);
        assertThat(response.status()).isEqualTo(KycStatus.VERIFIED);
    }

    // -------------------------------------------------------------------------
    // Test 4: verifyKyc sets verifiedAt and expiresAt exactly 2 years later
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("verifyKyc sets verifiedAt and expiresAt 2 years in the future")
    void verifyKyc_setsVerifiedAtAndExpiresAtTwoYearsLater() {
        KycRecord record = buildKycRecord(kycId, customerId, KycStatus.IN_PROGRESS, "US");

        when(kycRecordRepository.findById(kycId)).thenReturn(Optional.of(record));
        when(kycRecordRepository.save(any(KycRecord.class))).thenAnswer(inv -> inv.getArgument(0));

        kycAmlService.verifyKyc(kycId);

        ArgumentCaptor<KycRecord> captor = ArgumentCaptor.forClass(KycRecord.class);
        verify(kycRecordRepository).save(captor.capture());

        KycRecord saved = captor.getValue();
        assertThat(saved.getStatus()).isEqualTo(KycStatus.VERIFIED);
        assertThat(saved.getVerifiedAt()).isNotNull();
        assertThat(saved.getExpiresAt()).isNotNull();
        // expiresAt should be within a second of verifiedAt + 2 years
        assertThat(saved.getExpiresAt())
            .isAfterOrEqualTo(saved.getVerifiedAt().plusYears(2).minusSeconds(1))
            .isBeforeOrEqualTo(saved.getVerifiedAt().plusYears(2).plusSeconds(1));
    }

    // -------------------------------------------------------------------------
    // Test 5: rejectKyc sets status to REJECTED and stores the reason
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("rejectKyc transitions to REJECTED and persists rejection reason")
    void rejectKyc_setsRejectedStatusAndReason() {
        KycRecord record = buildKycRecord(kycId, customerId, KycStatus.IN_PROGRESS, "CA");

        when(kycRecordRepository.findById(kycId)).thenReturn(Optional.of(record));
        when(kycRecordRepository.save(any(KycRecord.class))).thenAnswer(inv -> inv.getArgument(0));

        kycAmlService.rejectKyc(kycId, "Document appears to be forged");

        ArgumentCaptor<KycRecord> captor = ArgumentCaptor.forClass(KycRecord.class);
        verify(kycRecordRepository).save(captor.capture());

        KycRecord saved = captor.getValue();
        assertThat(saved.getStatus()).isEqualTo(KycStatus.REJECTED);
        assertThat(saved.getRejectionReason()).isEqualTo("Document appears to be forged");
    }

    // -------------------------------------------------------------------------
    // Test 6: verifyKyc — high-risk nationality produces HIGH risk level
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("verifyKyc assigns HIGH riskLevel for customer with high-risk nationality")
    void verifyKyc_assignsHighRiskLevel_forHighRiskNationality() {
        KycRecord record = buildKycRecord(kycId, customerId, KycStatus.IN_PROGRESS, "IR"); // Iran

        when(kycRecordRepository.findById(kycId)).thenReturn(Optional.of(record));
        when(kycRecordRepository.save(any(KycRecord.class))).thenAnswer(inv -> inv.getArgument(0));

        kycAmlService.verifyKyc(kycId);

        ArgumentCaptor<KycRecord> captor = ArgumentCaptor.forClass(KycRecord.class);
        verify(kycRecordRepository).save(captor.capture());

        assertThat(captor.getValue().getRiskLevel()).isEqualTo(RiskLevel.HIGH);
    }

    // -------------------------------------------------------------------------
    // Test 7: runAmlCheck — CLEAR result for low risk score
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("runAmlCheck returns CLEAR result when simulated risk score is below 40")
    void runAmlCheck_returnsClear_whenRiskScoreBelow40() {
        // Choose a UUID whose MSB produces a score < 40 for PEP check (0-60 range)
        // We test the scoreToResult helper directly instead of relying on UUID hash
        String result = kycAmlService.scoreToResult(25);
        assertThat(result).isEqualTo("CLEAR");

        // Also verify end-to-end via mock
        RunAmlCheckRequest request = new RunAmlCheckRequest(customerId, "PEP");
        AmlCheck savedCheck = buildAmlCheck(amlId, customerId, "PEP", "CLEAR", 25);

        when(amlCheckRepository.save(any(AmlCheck.class))).thenReturn(savedCheck);

        AmlCheckResponse response = kycAmlService.runAmlCheck(request);
        assertThat(response.result()).isEqualTo("CLEAR");
        assertThat(response.riskScore()).isEqualTo(25);
    }

    // -------------------------------------------------------------------------
    // Test 8: runAmlCheck — FLAGGED result for high risk score
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("runAmlCheck returns FLAGGED result when simulated risk score is above 70")
    void runAmlCheck_returnsFlagged_whenRiskScoreAbove70() {
        String result = kycAmlService.scoreToResult(85);
        assertThat(result).isEqualTo("FLAGGED");

        RunAmlCheckRequest request = new RunAmlCheckRequest(customerId, "SANCTIONS");
        AmlCheck savedCheck = buildAmlCheck(amlId, customerId, "SANCTIONS", "FLAGGED", 85);

        when(amlCheckRepository.save(any(AmlCheck.class))).thenReturn(savedCheck);

        AmlCheckResponse response = kycAmlService.runAmlCheck(request);
        assertThat(response.result()).isEqualTo("FLAGGED");
    }

    // -------------------------------------------------------------------------
    // Test 9: runAmlCheck — REVIEW_REQUIRED result for mid-range risk score
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("runAmlCheck returns REVIEW_REQUIRED result when risk score is between 40 and 70")
    void runAmlCheck_returnsReviewRequired_whenRiskScoreMidRange() {
        String result = kycAmlService.scoreToResult(55);
        assertThat(result).isEqualTo("REVIEW_REQUIRED");

        RunAmlCheckRequest request = new RunAmlCheckRequest(customerId, "ADVERSE_MEDIA");
        AmlCheck savedCheck = buildAmlCheck(amlId, customerId, "ADVERSE_MEDIA", "REVIEW_REQUIRED", 55);

        when(amlCheckRepository.save(any(AmlCheck.class))).thenReturn(savedCheck);

        AmlCheckResponse response = kycAmlService.runAmlCheck(request);
        assertThat(response.result()).isEqualTo("REVIEW_REQUIRED");
    }

    // -------------------------------------------------------------------------
    // Test 10: resolveAmlCheck — sets resolution, resolvedBy, resolvedAt
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("resolveAmlCheck persists resolution details and timestamp")
    void resolveAmlCheck_persistsResolutionDetails() {
        AmlCheck check = buildAmlCheck(amlId, customerId, "SANCTIONS", "FLAGGED", 80);
        check.setResolvedAt(null); // unresolved

        when(amlCheckRepository.findById(amlId)).thenReturn(Optional.of(check));
        when(amlCheckRepository.save(any(AmlCheck.class))).thenAnswer(inv -> inv.getArgument(0));

        AmlCheckResponse response = kycAmlService.resolveAmlCheck(
            amlId, "Reviewed and cleared after additional documentation", "compliance-officer-1");

        ArgumentCaptor<AmlCheck> captor = ArgumentCaptor.forClass(AmlCheck.class);
        verify(amlCheckRepository).save(captor.capture());

        AmlCheck saved = captor.getValue();
        assertThat(saved.getResolution()).isEqualTo("Reviewed and cleared after additional documentation");
        assertThat(saved.getResolvedBy()).isEqualTo("compliance-officer-1");
        assertThat(saved.getResolvedAt()).isNotNull();
    }

    // -------------------------------------------------------------------------
    // Test 11: detectExpiredKyc — marks VERIFIED records past expiresAt as EXPIRED
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("detectExpiredKyc marks all found VERIFIED-expired records as EXPIRED")
    void detectExpiredKyc_marksExpiredRecords() {
        KycRecord record1 = buildKycRecord(UUID.randomUUID(), UUID.randomUUID(), KycStatus.VERIFIED, "US");
        record1.setExpiresAt(LocalDateTime.now().minusDays(10));
        KycRecord record2 = buildKycRecord(UUID.randomUUID(), UUID.randomUUID(), KycStatus.VERIFIED, "GB");
        record2.setExpiresAt(LocalDateTime.now().minusDays(5));

        when(kycRecordRepository.findByExpiresAtBeforeAndStatus(any(LocalDateTime.class), eq(KycStatus.VERIFIED)))
            .thenReturn(List.of(record1, record2));
        when(kycRecordRepository.saveAll(anyList())).thenAnswer(inv -> inv.getArgument(0));

        int count = kycAmlService.detectExpiredKyc();

        assertThat(count).isEqualTo(2);
        assertThat(record1.getStatus()).isEqualTo(KycStatus.EXPIRED);
        assertThat(record2.getStatus()).isEqualTo(KycStatus.EXPIRED);
        verify(kycRecordRepository).saveAll(anyList());
    }

    // -------------------------------------------------------------------------
    // Test 12: suspendKyc — transitions VERIFIED → SUSPENDED
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("suspendKyc transitions VERIFIED record to SUSPENDED")
    void suspendKyc_transitionsVerifiedToSuspended() {
        KycRecord record = buildKycRecord(kycId, customerId, KycStatus.VERIFIED, "AU");

        when(kycRecordRepository.findById(kycId)).thenReturn(Optional.of(record));
        when(kycRecordRepository.save(any(KycRecord.class))).thenAnswer(inv -> inv.getArgument(0));

        kycAmlService.suspendKyc(kycId);

        ArgumentCaptor<KycRecord> captor = ArgumentCaptor.forClass(KycRecord.class);
        verify(kycRecordRepository).save(captor.capture());
        assertThat(captor.getValue().getStatus()).isEqualTo(KycStatus.SUSPENDED);
    }

    // -------------------------------------------------------------------------
    // Test 13: suspendKyc — throws when record is not VERIFIED
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("suspendKyc throws InvalidStateTransitionException when record is not VERIFIED")
    void suspendKyc_throwsInvalidState_whenNotVerified() {
        KycRecord record = buildKycRecord(kycId, customerId, KycStatus.PENDING, "AU");

        when(kycRecordRepository.findById(kycId)).thenReturn(Optional.of(record));

        assertThatThrownBy(() -> kycAmlService.suspendKyc(kycId))
            .isInstanceOf(InvalidStateTransitionException.class)
            .hasMessageContaining("PENDING");

        verify(kycRecordRepository, never()).save(any());
    }

    // -------------------------------------------------------------------------
    // Test 14: getKycRecord — throws ResourceNotFoundException for unknown id
    // -------------------------------------------------------------------------
    @Test
    @DisplayName("getKycRecord throws ResourceNotFoundException for unknown id")
    void getKycRecord_throwsNotFound_forUnknownId() {
        when(kycRecordRepository.findById(kycId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> kycAmlService.getKycRecord(kycId))
            .isInstanceOf(ResourceNotFoundException.class)
            .hasMessageContaining(kycId.toString());
    }

    // -------------------------------------------------------------------------
    // Helper builders
    // -------------------------------------------------------------------------

    private KycRecord buildKycRecord(UUID id, UUID custId, KycStatus status, String nationality) {
        KycRecord record = new KycRecord();
        record.setId(id);
        record.setCustomerId(custId);
        record.setFirstName("John");
        record.setLastName("Doe");
        record.setDateOfBirth(LocalDate.of(1985, 1, 1));
        record.setNationality(nationality);
        record.setDocumentType("PASSPORT");
        record.setDocumentNumber("AB123456");
        record.setDocumentExpiry(LocalDate.now().plusYears(5));
        record.setStatus(status);
        record.setCreatedAt(LocalDateTime.now());
        record.setUpdatedAt(LocalDateTime.now());
        return record;
    }

    private AmlCheck buildAmlCheck(UUID id, UUID custId, String checkType, String result, int score) {
        AmlCheck check = new AmlCheck();
        check.setId(id);
        check.setCustomerId(custId);
        check.setCheckType(checkType);
        check.setResult(result);
        check.setRiskScore(score);
        check.setDetails("Simulated check details");
        check.setCheckedAt(LocalDateTime.now());
        return check;
    }
}
