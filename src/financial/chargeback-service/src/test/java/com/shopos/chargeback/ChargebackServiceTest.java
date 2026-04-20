package com.shopos.chargeback;

import com.shopos.chargeback.domain.Chargeback;
import com.shopos.chargeback.domain.ChargebackStatus;
import com.shopos.chargeback.dto.ChargebackResponse;
import com.shopos.chargeback.dto.CreateChargebackRequest;
import com.shopos.chargeback.repository.ChargebackRepository;
import com.shopos.chargeback.service.ChargebackService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ChargebackServiceTest {

    @Mock
    private ChargebackRepository chargebackRepository;

    @InjectMocks
    private ChargebackService chargebackService;

    private Chargeback sampleChargeback;

    @BeforeEach
    void setUp() {
        sampleChargeback = Chargeback.builder()
                .id(UUID.randomUUID())
                .paymentId("pay-001")
                .orderId("ord-001")
                .customerId("cust-001")
                .amount(new BigDecimal("150.00"))
                .currency("USD")
                .status(ChargebackStatus.OPEN)
                .build();
        sampleChargeback.onCreate();
    }

    @Test
    void createChargeback_shouldReturnOpenStatus() {
        when(chargebackRepository.save(any(Chargeback.class))).thenReturn(sampleChargeback);

        CreateChargebackRequest request = new CreateChargebackRequest();
        request.setPaymentId("pay-001");
        request.setOrderId("ord-001");
        request.setCustomerId("cust-001");
        request.setAmount(new BigDecimal("150.00"));
        request.setCurrency("USD");

        ChargebackResponse response = chargebackService.createChargeback(request);

        assertThat(response.getStatus()).isEqualTo(ChargebackStatus.OPEN);
        assertThat(response.getPaymentId()).isEqualTo("pay-001");
    }

    @Test
    void getChargeback_shouldReturnCorrectData() {
        when(chargebackRepository.findById(sampleChargeback.getId()))
                .thenReturn(Optional.of(sampleChargeback));

        ChargebackResponse response = chargebackService.getChargeback(sampleChargeback.getId());

        assertThat(response.getId()).isEqualTo(sampleChargeback.getId());
        assertThat(response.getCustomerId()).isEqualTo("cust-001");
    }
}
