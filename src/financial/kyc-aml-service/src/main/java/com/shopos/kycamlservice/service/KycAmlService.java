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
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Random;
import java.util.Set;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class KycAmlService {

    /**
     * ISO 3166-1 alpha-2 codes designated as high-risk jurisdictions.
     * IR = Iran, KP = North Korea, SY = Syria, CU = Cuba, SD = Sudan
     */
    private static final Set<String> HIGH_RISK_COUNTRIES = Set.of("IR", "KP", "SY", "CU", "SD");

    /**
     * KYC records are valid for two years after verification.
     */
    private static final int KYC_VALIDITY_YEARS = 2;

    private final KycRecordRepository kycRecordRepository;
    private final AmlCheckRepository amlCheckRepository;
    private final Random random = new Random();

    // -------------------------------------------------------------------------
    // KYC operations
    // -------------------------------------------------------------------------

    @Transactional
    public KycResponse createKycRecord(CreateKycRequest request) {
        if (kycRecordRepository.existsByCustomerId(request.customerId())) {
            throw new ConflictException(
                "A KYC record already exists for customerId: " + request.customerId());
        }

        KycRecord record = KycRecord.builder()
            .customerId(request.customerId())
            .firstName(request.firstName())
            .lastName(request.lastName())
            .dateOfBirth(request.dateOfBirth())
            .nationality(request.nationality().toUpperCase())
            .documentType(request.documentType())
            .documentNumber(request.documentNumber())
            .documentExpiry(request.documentExpiry())
            .status(KycStatus.PENDING)
            .notes(request.notes())
            .build();

        KycRecord saved = kycRecordRepository.save(record);
        log.info("Created KYC record id={} for customerId={}", saved.getId(), saved.getCustomerId());
        return KycResponse.from(saved);
    }

    @Transactional(readOnly = true)
    public KycResponse getKycRecord(UUID id) {
        KycRecord record = findKycById(id);
        return KycResponse.from(record);
    }

    @Transactional(readOnly = true)
    public KycResponse getByCustomerId(UUID customerId) {
        KycRecord record = kycRecordRepository.findByCustomerId(customerId)
            .orElseThrow(() -> new ResourceNotFoundException(
                "KYC record not found for customerId: " + customerId));
        return KycResponse.from(record);
    }

    @Transactional
    public void startVerification(UUID id) {
        KycRecord record = findKycById(id);
        if (record.getStatus() != KycStatus.PENDING) {
            throw new InvalidStateTransitionException(
                "Cannot start verification: current status is " + record.getStatus()
                    + ". Expected PENDING.");
        }
        record.setStatus(KycStatus.IN_PROGRESS);
        kycRecordRepository.save(record);
        log.info("KYC verification started for id={}", id);
    }

    @Transactional
    public void verifyKyc(UUID id) {
        KycRecord record = findKycById(id);
        if (record.getStatus() != KycStatus.IN_PROGRESS) {
            throw new InvalidStateTransitionException(
                "Cannot verify KYC: current status is " + record.getStatus()
                    + ". Expected IN_PROGRESS.");
        }

        LocalDateTime now = LocalDateTime.now();
        RiskLevel riskLevel = computeRiskLevel(record.getNationality());

        record.setStatus(KycStatus.VERIFIED);
        record.setRiskLevel(riskLevel);
        record.setVerifiedAt(now);
        record.setExpiresAt(now.plusYears(KYC_VALIDITY_YEARS));

        kycRecordRepository.save(record);
        log.info("KYC verified for id={} riskLevel={} expiresAt={}", id, riskLevel, record.getExpiresAt());
    }

    @Transactional
    public void rejectKyc(UUID id, String reason) {
        KycRecord record = findKycById(id);
        if (record.getStatus() != KycStatus.IN_PROGRESS) {
            throw new InvalidStateTransitionException(
                "Cannot reject KYC: current status is " + record.getStatus()
                    + ". Expected IN_PROGRESS.");
        }
        record.setStatus(KycStatus.REJECTED);
        record.setRejectionReason(reason);
        kycRecordRepository.save(record);
        log.info("KYC rejected for id={} reason={}", id, reason);
    }

    @Transactional
    public void suspendKyc(UUID id) {
        KycRecord record = findKycById(id);
        if (record.getStatus() != KycStatus.VERIFIED) {
            throw new InvalidStateTransitionException(
                "Cannot suspend KYC: current status is " + record.getStatus()
                    + ". Expected VERIFIED.");
        }
        record.setStatus(KycStatus.SUSPENDED);
        kycRecordRepository.save(record);
        log.info("KYC suspended for id={}", id);
    }

    /**
     * Batch job: marks VERIFIED records whose expiresAt is in the past as EXPIRED.
     * Runs daily at 01:00 UTC.
     *
     * @return number of records marked expired
     */
    @Scheduled(cron = "0 0 1 * * *")
    @Transactional
    public int detectExpiredKyc() {
        LocalDateTime now = LocalDateTime.now();
        List<KycRecord> expiredRecords =
            kycRecordRepository.findByExpiresAtBeforeAndStatus(now, KycStatus.VERIFIED);

        for (KycRecord record : expiredRecords) {
            record.setStatus(KycStatus.EXPIRED);
        }
        kycRecordRepository.saveAll(expiredRecords);

        if (!expiredRecords.isEmpty()) {
            log.info("Marked {} KYC records as EXPIRED", expiredRecords.size());
        }
        return expiredRecords.size();
    }

    // -------------------------------------------------------------------------
    // AML operations
    // -------------------------------------------------------------------------

    @Transactional
    public AmlCheckResponse runAmlCheck(RunAmlCheckRequest request) {
        int riskScore = simulateRiskScore(request.checkType(), request.customerId());
        String result = scoreToResult(riskScore);

        String details = buildCheckDetails(request.checkType(), riskScore, result);

        AmlCheck check = AmlCheck.builder()
            .customerId(request.customerId())
            .checkType(request.checkType())
            .result(result)
            .riskScore(riskScore)
            .details(details)
            .checkedAt(LocalDateTime.now())
            .build();

        AmlCheck saved = amlCheckRepository.save(check);
        log.info("AML check id={} customerId={} type={} result={} score={}",
            saved.getId(), saved.getCustomerId(), saved.getCheckType(),
            saved.getResult(), saved.getRiskScore());
        return AmlCheckResponse.from(saved);
    }

    @Transactional(readOnly = true)
    public List<AmlCheckResponse> getAmlChecks(UUID customerId) {
        return amlCheckRepository.findByCustomerId(customerId)
            .stream()
            .map(AmlCheckResponse::from)
            .toList();
    }

    @Transactional
    public AmlCheckResponse resolveAmlCheck(UUID id, String resolution, String resolvedBy) {
        AmlCheck check = amlCheckRepository.findById(id)
            .orElseThrow(() -> new ResourceNotFoundException("AML check not found with id: " + id));

        if (check.getResolvedAt() != null) {
            throw new InvalidStateTransitionException(
                "AML check id=" + id + " has already been resolved.");
        }

        check.setResolution(resolution);
        check.setResolvedBy(resolvedBy);
        check.setResolvedAt(LocalDateTime.now());

        AmlCheck saved = amlCheckRepository.save(check);
        log.info("AML check id={} resolved by={}", id, resolvedBy);
        return AmlCheckResponse.from(saved);
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    private KycRecord findKycById(UUID id) {
        return kycRecordRepository.findById(id)
            .orElseThrow(() -> new ResourceNotFoundException("KYC record not found with id: " + id));
    }

    /**
     * Determines risk level based on nationality.
     * Citizens of sanctioned / high-risk countries are rated HIGH.
     * All others are rated LOW by default (upgraded by AML checks later).
     */
    private RiskLevel computeRiskLevel(String nationality) {
        if (nationality != null && HIGH_RISK_COUNTRIES.contains(nationality.toUpperCase())) {
            return RiskLevel.HIGH;
        }
        return RiskLevel.LOW;
    }

    /**
     * Simulates an AML risk score.
     *
     * <p>In production this would call a third-party screening API (e.g., Refinitiv World-Check,
     * ComplyAdvantage). Here we use deterministic-ish mock logic so that tests can rely on
     * predictable behaviour by seeding the customerId.
     *
     * <ul>
     *   <li>SANCTIONS checks use the most significant byte of the customer UUID as the base score,
     *       producing a wide spread (0–255 clamped to 0–100).</li>
     *   <li>PEP checks use a lighter scoring curve (0–60).</li>
     *   <li>ADVERSE_MEDIA checks use a random 0–50 range.</li>
     *   <li>TRANSACTION_MONITORING uses a random 0–80 range.</li>
     * </ul>
     */
    int simulateRiskScore(String checkType, UUID customerId) {
        // Seed from the most-significant long of the UUID for reproducibility in tests
        long seed = customerId.getMostSignificantBits();
        Random seededRandom = new Random(seed);

        return switch (checkType) {
            case "SANCTIONS" -> {
                // Score 0-100 with a bias toward higher values for flagging demos
                int base = (int) (Math.abs(seed) % 101);
                yield Math.min(100, Math.max(0, base));
            }
            case "PEP" -> seededRandom.nextInt(61);          // 0-60
            case "ADVERSE_MEDIA" -> seededRandom.nextInt(51); // 0-50
            case "TRANSACTION_MONITORING" -> seededRandom.nextInt(81); // 0-80
            default -> seededRandom.nextInt(101);
        };
    }

    /**
     * Maps a numeric risk score to an AML result string.
     *
     * <ul>
     *   <li>score > 70  → FLAGGED</li>
     *   <li>score 40-70 → REVIEW_REQUIRED</li>
     *   <li>score < 40  → CLEAR</li>
     * </ul>
     */
    String scoreToResult(int score) {
        if (score > 70) {
            return "FLAGGED";
        } else if (score >= 40) {
            return "REVIEW_REQUIRED";
        } else {
            return "CLEAR";
        }
    }

    private String buildCheckDetails(String checkType, int riskScore, String result) {
        return String.format(
            "Automated %s screening completed. Risk score: %d/100. Outcome: %s. "
                + "This result was produced by the internal simulation engine and requires "
                + "manual review for FLAGGED or REVIEW_REQUIRED outcomes.",
            checkType, riskScore, result);
    }
}
