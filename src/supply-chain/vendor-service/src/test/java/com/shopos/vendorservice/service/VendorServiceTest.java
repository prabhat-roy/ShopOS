package com.shopos.vendorservice.service;

import com.shopos.vendorservice.domain.Vendor;
import com.shopos.vendorservice.domain.VendorStatus;
import com.shopos.vendorservice.dto.CreateVendorRequest;
import com.shopos.vendorservice.dto.VendorResponse;
import com.shopos.vendorservice.exception.VendorConflictException;
import com.shopos.vendorservice.exception.VendorNotFoundException;
import com.shopos.vendorservice.repository.VendorRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class VendorServiceTest {

    @Mock
    private VendorRepository vendorRepository;

    @InjectMocks
    private VendorService vendorService;

    private Vendor sampleVendor;
    private UUID vendorId;

    @BeforeEach
    void setUp() {
        vendorId = UUID.randomUUID();
        sampleVendor = Vendor.builder()
                .id(vendorId)
                .name("Acme Supplies Ltd")
                .email("contact@acme.com")
                .phone("+1-800-123-4567")
                .website("https://acme.com")
                .status(VendorStatus.ACTIVE)
                .country("US")
                .address("123 Main St, New York, NY 10001")
                .taxId("US-12345678")
                .rating(new BigDecimal("4.50"))
                .totalOrders(42)
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .build();
    }

    @Test
    @DisplayName("createVendor - success: returns VendorResponse with PENDING_APPROVAL status")
    void createVendor_success() {
        CreateVendorRequest request = new CreateVendorRequest(
                "Acme Supplies Ltd",
                "contact@acme.com",
                "+1-800-123-4567",
                "https://acme.com",
                "US",
                "123 Main St",
                "US-12345678"
        );

        when(vendorRepository.existsByEmail("contact@acme.com")).thenReturn(false);
        when(vendorRepository.save(any(Vendor.class))).thenAnswer(invocation -> {
            Vendor v = invocation.getArgument(0);
            v.setId(vendorId);
            v.setCreatedAt(Instant.now());
            v.setUpdatedAt(Instant.now());
            return v;
        });

        VendorResponse response = vendorService.createVendor(request);

        assertThat(response).isNotNull();
        assertThat(response.name()).isEqualTo("Acme Supplies Ltd");
        assertThat(response.email()).isEqualTo("contact@acme.com");
        assertThat(response.status()).isEqualTo(VendorStatus.PENDING_APPROVAL);
        verify(vendorRepository).existsByEmail("contact@acme.com");
        verify(vendorRepository).save(any(Vendor.class));
    }

    @Test
    @DisplayName("createVendor - duplicate email: throws VendorConflictException")
    void createVendor_duplicateEmail_throwsConflict() {
        CreateVendorRequest request = new CreateVendorRequest(
                "Other Vendor",
                "contact@acme.com",
                "+1-800-999-0000",
                null, null, null, null
        );

        when(vendorRepository.existsByEmail("contact@acme.com")).thenReturn(true);

        assertThatThrownBy(() -> vendorService.createVendor(request))
                .isInstanceOf(VendorConflictException.class)
                .hasMessageContaining("contact@acme.com");

        verify(vendorRepository).existsByEmail("contact@acme.com");
        verify(vendorRepository, never()).save(any());
    }

    @Test
    @DisplayName("getVendor - found: returns correct VendorResponse")
    void getVendor_found() {
        when(vendorRepository.findById(vendorId)).thenReturn(Optional.of(sampleVendor));

        VendorResponse response = vendorService.getVendor(vendorId);

        assertThat(response.id()).isEqualTo(vendorId);
        assertThat(response.name()).isEqualTo("Acme Supplies Ltd");
        assertThat(response.email()).isEqualTo("contact@acme.com");
        assertThat(response.status()).isEqualTo(VendorStatus.ACTIVE);
        verify(vendorRepository).findById(vendorId);
    }

    @Test
    @DisplayName("getVendor - not found: throws VendorNotFoundException")
    void getVendor_notFound_throwsException() {
        UUID unknownId = UUID.randomUUID();
        when(vendorRepository.findById(unknownId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> vendorService.getVendor(unknownId))
                .isInstanceOf(VendorNotFoundException.class)
                .hasMessageContaining(unknownId.toString());

        verify(vendorRepository).findById(unknownId);
    }

    @Test
    @DisplayName("listVendors - no filter: returns all vendors")
    void listVendors_noFilter_returnsAll() {
        Vendor second = Vendor.builder()
                .id(UUID.randomUUID())
                .name("Beta Corp")
                .email("info@beta.com")
                .phone("+44-20-1234-5678")
                .status(VendorStatus.INACTIVE)
                .totalOrders(0)
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .build();

        when(vendorRepository.findAll()).thenReturn(List.of(sampleVendor, second));

        List<VendorResponse> result = vendorService.listVendors(null);

        assertThat(result).hasSize(2);
        verify(vendorRepository).findAll();
        verify(vendorRepository, never()).findByStatus(any());
    }

    @Test
    @DisplayName("listVendors - by status: delegates to findByStatus")
    void listVendors_byStatus_filtersCorrectly() {
        when(vendorRepository.findByStatus(VendorStatus.ACTIVE)).thenReturn(List.of(sampleVendor));

        List<VendorResponse> result = vendorService.listVendors(VendorStatus.ACTIVE);

        assertThat(result).hasSize(1);
        assertThat(result.get(0).status()).isEqualTo(VendorStatus.ACTIVE);
        verify(vendorRepository).findByStatus(VendorStatus.ACTIVE);
        verify(vendorRepository, never()).findAll();
    }

    @Test
    @DisplayName("suspendVendor - existing vendor: sets status to SUSPENDED")
    void suspendVendor_setsStatusToSuspended() {
        when(vendorRepository.findById(vendorId)).thenReturn(Optional.of(sampleVendor));
        when(vendorRepository.save(any(Vendor.class))).thenReturn(sampleVendor);

        vendorService.suspendVendor(vendorId);

        assertThat(sampleVendor.getStatus()).isEqualTo(VendorStatus.SUSPENDED);
        verify(vendorRepository).findById(vendorId);
        verify(vendorRepository).save(sampleVendor);
    }

    @Test
    @DisplayName("activateVendor - existing vendor: sets status to ACTIVE")
    void activateVendor_setsStatusToActive() {
        sampleVendor.setStatus(VendorStatus.SUSPENDED);
        when(vendorRepository.findById(vendorId)).thenReturn(Optional.of(sampleVendor));
        when(vendorRepository.save(any(Vendor.class))).thenReturn(sampleVendor);

        vendorService.activateVendor(vendorId);

        assertThat(sampleVendor.getStatus()).isEqualTo(VendorStatus.ACTIVE);
        verify(vendorRepository).findById(vendorId);
        verify(vendorRepository).save(sampleVendor);
    }
}
