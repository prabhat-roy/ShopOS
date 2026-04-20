package com.shopos.payoutservice.service;

import com.shopos.payoutservice.domain.Payout;
import com.shopos.payoutservice.domain.PayoutMethod;
import com.shopos.payoutservice.domain.PayoutStatus;
import com.shopos.payoutservice.dto.CreatePayoutRequest;
import com.shopos.payoutservice.dto.PayoutResponse;
import com.shopos.payoutservice.repository.PayoutRepository;
import jakarta.persistence.EntityNotFoundException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.atLeastOnce;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@DisplayName("PayoutService unit tests")
class PayoutServiceTest {

    @Mock
    private PayoutRepository payoutRepository;

    @InjectMocks
    private PayoutService payoutService;

    private UUID payoutId;
    private Payout pendingPayout;

    @BeforeEach
    void setUp() {
        payoutId = UUID.randomUUID();
        pendingPayout = Payout.builder()
                .id(payoutId)
                .vendorId(UUID.randomUUID())
                .amount(new BigDecimal("500.00"))
                .currency("USD")
                .status(PayoutStatus.PENDING)
                .method(PayoutMethod.ACH)
                .reference("PAY-ABCD1234")
                .bankAccount("{\"accountNumber\":\"12345678\",\"routingNumber\":\"021000021\"}")
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();
    }

    // -----------------------------------------------------------------------
    // Test 1 — createPayout generates a properly formatted reference
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("createPayout — should generate PAY-XXXXXXXX reference")
    void createPayout_generatesCorrectReference() {
        CreatePayoutRequest request = buildCreateRequest();
        when(payoutRepository.save(any(Payout.class))).thenAnswer(inv -> inv.getArgument(0));

        PayoutResponse response = payoutService.createPayout(request);

        assertThat(response.reference()).matches("PAY-[A-F0-9]{8}");
    }

    // -----------------------------------------------------------------------
    // Test 2 — createPayout persists PENDING status
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("createPayout — new payout must be in PENDING status")
    void createPayout_statusIsPending() {
        CreatePayoutRequest request = buildCreateRequest();
        when(payoutRepository.save(any(Payout.class))).thenAnswer(inv -> inv.getArgument(0));

        PayoutResponse response = payoutService.createPayout(request);

        assertThat(response.status()).isEqualTo(PayoutStatus.PENDING);
    }

    // -----------------------------------------------------------------------
    // Test 3 — getPayout throws EntityNotFoundException for unknown id
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("getPayout — throws EntityNotFoundException when payout not found")
    void getPayout_throwsWhenNotFound() {
        UUID unknownId = UUID.randomUUID();
        when(payoutRepository.findById(unknownId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> payoutService.getPayout(unknownId))
                .isInstanceOf(EntityNotFoundException.class)
                .hasMessageContaining(unknownId.toString());
    }

    // -----------------------------------------------------------------------
    // Test 4 — processPayout transitions PENDING → COMPLETED and sets processedAt
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("processPayout — transitions PENDING to COMPLETED and sets processedAt")
    void processPayout_pendingToCompleted() {
        when(payoutRepository.findById(payoutId)).thenReturn(Optional.of(pendingPayout));
        when(payoutRepository.save(any(Payout.class))).thenAnswer(inv -> inv.getArgument(0));

        payoutService.processPayout(payoutId);

        ArgumentCaptor<Payout> captor = ArgumentCaptor.forClass(Payout.class);
        verify(payoutRepository, atLeastOnce()).save(captor.capture());

        // The last saved state should be COMPLETED
        List<Payout> allSaved = captor.getAllValues();
        Payout lastSaved = allSaved.get(allSaved.size() - 1);
        assertThat(lastSaved.getStatus()).isEqualTo(PayoutStatus.COMPLETED);
        assertThat(lastSaved.getProcessedAt()).isNotNull();
    }

    // -----------------------------------------------------------------------
    // Test 5 — processPayout throws when payout is not PENDING
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("processPayout — throws IllegalStateException when payout is not PENDING")
    void processPayout_throwsWhenNotPending() {
        pendingPayout.setStatus(PayoutStatus.COMPLETED);
        when(payoutRepository.findById(payoutId)).thenReturn(Optional.of(pendingPayout));

        assertThatThrownBy(() -> payoutService.processPayout(payoutId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("PENDING");
    }

    // -----------------------------------------------------------------------
    // Test 6 — failPayout sets FAILED status and records failure reason
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("failPayout — sets FAILED status and records failure reason")
    void failPayout_setsFailedWithReason() {
        pendingPayout.setStatus(PayoutStatus.PROCESSING);
        when(payoutRepository.findById(payoutId)).thenReturn(Optional.of(pendingPayout));
        when(payoutRepository.save(any(Payout.class))).thenAnswer(inv -> inv.getArgument(0));

        payoutService.failPayout(payoutId, "Insufficient funds");

        ArgumentCaptor<Payout> captor = ArgumentCaptor.forClass(Payout.class);
        verify(payoutRepository).save(captor.capture());
        Payout saved = captor.getValue();
        assertThat(saved.getStatus()).isEqualTo(PayoutStatus.FAILED);
        assertThat(saved.getFailureReason()).isEqualTo("Insufficient funds");
    }

    // -----------------------------------------------------------------------
    // Test 7 — cancelPayout transitions PENDING → CANCELLED
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("cancelPayout — transitions PENDING to CANCELLED")
    void cancelPayout_pendingToCancelled() {
        when(payoutRepository.findById(payoutId)).thenReturn(Optional.of(pendingPayout));
        when(payoutRepository.save(any(Payout.class))).thenAnswer(inv -> inv.getArgument(0));

        payoutService.cancelPayout(payoutId);

        ArgumentCaptor<Payout> captor = ArgumentCaptor.forClass(Payout.class);
        verify(payoutRepository).save(captor.capture());
        assertThat(captor.getValue().getStatus()).isEqualTo(PayoutStatus.CANCELLED);
    }

    // -----------------------------------------------------------------------
    // Test 8 — cancelPayout throws when payout is not PENDING
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("cancelPayout — throws IllegalStateException when payout is COMPLETED")
    void cancelPayout_throwsWhenNotPending() {
        pendingPayout.setStatus(PayoutStatus.COMPLETED);
        when(payoutRepository.findById(payoutId)).thenReturn(Optional.of(pendingPayout));

        assertThatThrownBy(() -> payoutService.cancelPayout(payoutId))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("PENDING");
    }

    // -----------------------------------------------------------------------
    // Test 9 — retryPayout resets FAILED → PENDING and clears failureReason
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("retryPayout — resets FAILED to PENDING and clears failure reason")
    void retryPayout_failedToPending() {
        pendingPayout.setStatus(PayoutStatus.FAILED);
        pendingPayout.setFailureReason("Network timeout");
        when(payoutRepository.findById(payoutId)).thenReturn(Optional.of(pendingPayout));
        when(payoutRepository.save(any(Payout.class))).thenAnswer(inv -> inv.getArgument(0));

        payoutService.retryPayout(payoutId);

        ArgumentCaptor<Payout> captor = ArgumentCaptor.forClass(Payout.class);
        verify(payoutRepository).save(captor.capture());
        Payout saved = captor.getValue();
        assertThat(saved.getStatus()).isEqualTo(PayoutStatus.PENDING);
        assertThat(saved.getFailureReason()).isNull();
    }

    // -----------------------------------------------------------------------
    // Test 10 — processDuePayouts processes all due PENDING payouts
    // -----------------------------------------------------------------------
    @Test
    @DisplayName("processDuePayouts — processes all due PENDING payouts and returns count")
    void processDuePayouts_processesAndReturnsCount() {
        Payout due1 = buildPayout(PayoutStatus.PENDING);
        Payout due2 = buildPayout(PayoutStatus.PENDING);
        when(payoutRepository.findDuePayouts(any(LocalDateTime.class)))
                .thenReturn(List.of(due1, due2));
        when(payoutRepository.save(any(Payout.class))).thenAnswer(inv -> inv.getArgument(0));

        int count = payoutService.processDuePayouts();

        assertThat(count).isEqualTo(2);
    }

    // -----------------------------------------------------------------------
    // Helpers
    // -----------------------------------------------------------------------
    private CreatePayoutRequest buildCreateRequest() {
        return new CreatePayoutRequest(
                UUID.randomUUID(),
                new BigDecimal("250.00"),
                "USD",
                PayoutMethod.BANK_TRANSFER,
                "{\"accountNumber\":\"987654321\",\"routingNumber\":\"011000138\"}",
                null
        );
    }

    private Payout buildPayout(PayoutStatus status) {
        return Payout.builder()
                .id(UUID.randomUUID())
                .vendorId(UUID.randomUUID())
                .amount(new BigDecimal("100.00"))
                .currency("USD")
                .status(status)
                .method(PayoutMethod.ACH)
                .reference("PAY-" + UUID.randomUUID().toString().substring(0, 8).toUpperCase())
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();
    }
}
