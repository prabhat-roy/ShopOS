package com.shopos.payoutservice.service;

import com.shopos.payoutservice.domain.Payout;
import com.shopos.payoutservice.domain.PayoutStatus;
import com.shopos.payoutservice.dto.CreatePayoutRequest;
import com.shopos.payoutservice.dto.PayoutResponse;
import com.shopos.payoutservice.repository.PayoutRepository;
import jakarta.persistence.EntityNotFoundException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class PayoutService {

    private final PayoutRepository payoutRepository;

    /**
     * Creates a new payout in PENDING status and generates its unique reference.
     * Reference format: PAY-XXXXXXXX where XXXXXXXX is 8 uppercase hex chars of a UUID.
     */
    @Transactional
    public PayoutResponse createPayout(CreatePayoutRequest request) {
        String reference = generateReference();

        Payout payout = Payout.builder()
                .vendorId(request.vendorId())
                .amount(request.amount())
                .currency(request.currency() != null ? request.currency() : "USD")
                .status(PayoutStatus.PENDING)
                .method(request.method())
                .reference(reference)
                .bankAccount(request.bankAccount())
                .scheduledAt(request.scheduledAt())
                .build();

        Payout saved = payoutRepository.save(payout);
        log.info("Created payout {} for vendorId={} amount={} {}",
                saved.getReference(), saved.getVendorId(), saved.getAmount(), saved.getCurrency());
        return PayoutResponse.from(saved);
    }

    /**
     * Retrieves a single payout by its UUID.
     *
     * @throws EntityNotFoundException if no payout exists with that id
     */
    @Transactional(readOnly = true)
    public PayoutResponse getPayout(UUID id) {
        return PayoutResponse.from(findOrThrow(id));
    }

    /**
     * Returns a paginated list of payouts for a given vendor, optionally filtered by status.
     */
    @Transactional(readOnly = true)
    public Page<PayoutResponse> listByVendor(UUID vendorId, PayoutStatus status, Pageable pageable) {
        Page<Payout> page;
        if (status != null) {
            page = payoutRepository.findByVendorIdAndStatus(vendorId, status, pageable);
        } else {
            page = payoutRepository.findByVendorId(vendorId, pageable);
        }
        return page.map(PayoutResponse::from);
    }

    /**
     * Returns all payouts with the given status.
     */
    @Transactional(readOnly = true)
    public List<PayoutResponse> listByStatus(PayoutStatus status) {
        return payoutRepository.findByStatus(status)
                .stream()
                .map(PayoutResponse::from)
                .toList();
    }

    /**
     * Processes a PENDING payout: PENDING → PROCESSING → COMPLETED.
     * Sets processedAt on successful completion.
     *
     * @throws IllegalStateException if the payout is not in PENDING status
     */
    @Transactional
    public void processPayout(UUID id) {
        Payout payout = findOrThrow(id);
        requireStatus(payout, PayoutStatus.PENDING, "process");

        payout.setStatus(PayoutStatus.PROCESSING);
        payoutRepository.save(payout);
        log.info("Payout {} moved to PROCESSING", payout.getReference());

        // Simulate the payment-rail call. In production this would invoke
        // an async external adapter; here we complete synchronously.
        payout.setStatus(PayoutStatus.COMPLETED);
        payout.setProcessedAt(LocalDateTime.now());
        payout.setFailureReason(null);
        payoutRepository.save(payout);
        log.info("Payout {} COMPLETED at {}", payout.getReference(), payout.getProcessedAt());
    }

    /**
     * Marks a PROCESSING payout as FAILED and records the reason.
     *
     * @throws IllegalStateException if the payout is not in PROCESSING status
     */
    @Transactional
    public void failPayout(UUID id, String failureReason) {
        Payout payout = findOrThrow(id);
        requireStatus(payout, PayoutStatus.PROCESSING, "fail");
        payout.setStatus(PayoutStatus.FAILED);
        payout.setFailureReason(failureReason);
        payoutRepository.save(payout);
        log.warn("Payout {} FAILED: {}", payout.getReference(), failureReason);
    }

    /**
     * Cancels a PENDING payout.
     *
     * @throws IllegalStateException if the payout is not in PENDING status
     */
    @Transactional
    public void cancelPayout(UUID id) {
        Payout payout = findOrThrow(id);
        requireStatus(payout, PayoutStatus.PENDING, "cancel");
        payout.setStatus(PayoutStatus.CANCELLED);
        payoutRepository.save(payout);
        log.info("Payout {} CANCELLED", payout.getReference());
    }

    /**
     * Retries a FAILED payout by resetting it to PENDING.
     *
     * @throws IllegalStateException if the payout is not in FAILED status
     */
    @Transactional
    public void retryPayout(UUID id) {
        Payout payout = findOrThrow(id);
        requireStatus(payout, PayoutStatus.FAILED, "retry");
        payout.setStatus(PayoutStatus.PENDING);
        payout.setFailureReason(null);
        payoutRepository.save(payout);
        log.info("Payout {} reset to PENDING for retry", payout.getReference());
    }

    /**
     * Batch operation: processes all PENDING payouts whose scheduledAt is in the past
     * or have no schedule (immediately eligible).
     *
     * @return the number of payouts that were successfully processed
     */
    @Transactional
    public int processDuePayouts() {
        List<Payout> duePending = payoutRepository.findDuePayouts(LocalDateTime.now());
        int processed = 0;
        for (Payout payout : duePending) {
            try {
                payout.setStatus(PayoutStatus.PROCESSING);
                payoutRepository.save(payout);

                // Payment-rail invocation would go here in production.
                payout.setStatus(PayoutStatus.COMPLETED);
                payout.setProcessedAt(LocalDateTime.now());
                payout.setFailureReason(null);
                payoutRepository.save(payout);
                processed++;
                log.info("Batch processed payout {} for vendor {}",
                        payout.getReference(), payout.getVendorId());
            } catch (Exception e) {
                log.error("Batch processing failed for payout {}: {}",
                        payout.getReference(), e.getMessage());
                payout.setStatus(PayoutStatus.FAILED);
                payout.setFailureReason(e.getMessage());
                payoutRepository.save(payout);
            }
        }
        log.info("processDuePayouts: processed {} of {} due payouts", processed, duePending.size());
        return processed;
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    private Payout findOrThrow(UUID id) {
        return payoutRepository.findById(id)
                .orElseThrow(() -> new EntityNotFoundException(
                        "Payout not found with id: " + id));
    }

    private void requireStatus(Payout payout, PayoutStatus required, String operation) {
        if (payout.getStatus() != required) {
            throw new IllegalStateException(
                    "Cannot " + operation + " payout " + payout.getReference() +
                    ". Required status: " + required + ", actual: " + payout.getStatus());
        }
    }

    /**
     * Generates a unique payout reference: PAY-XXXXXXXX
     * where XXXXXXXX is the first 8 uppercase hex characters of a random UUID.
     */
    private String generateReference() {
        String hex = UUID.randomUUID().toString().replace("-", "").substring(0, 8).toUpperCase();
        return "PAY-" + hex;
    }
}
