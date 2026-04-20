package com.shopos.vendorservice.service;

import com.shopos.vendorservice.domain.Vendor;
import com.shopos.vendorservice.domain.VendorStatus;
import com.shopos.vendorservice.dto.CreateVendorRequest;
import com.shopos.vendorservice.dto.UpdateVendorRequest;
import com.shopos.vendorservice.dto.VendorResponse;
import com.shopos.vendorservice.exception.VendorConflictException;
import com.shopos.vendorservice.exception.VendorNotFoundException;
import com.shopos.vendorservice.repository.VendorRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class VendorService {

    private final VendorRepository vendorRepository;

    @Transactional
    public VendorResponse createVendor(CreateVendorRequest request) {
        if (vendorRepository.existsByEmail(request.email())) {
            throw new VendorConflictException(request.email());
        }
        Vendor vendor = Vendor.builder()
                .name(request.name())
                .email(request.email())
                .phone(request.phone())
                .website(request.website())
                .country(request.country())
                .address(request.address())
                .taxId(request.taxId())
                .status(VendorStatus.PENDING_APPROVAL)
                .totalOrders(0)
                .build();
        Vendor saved = vendorRepository.save(vendor);
        return VendorResponse.from(saved);
    }

    @Transactional(readOnly = true)
    public VendorResponse getVendor(UUID id) {
        Vendor vendor = vendorRepository.findById(id)
                .orElseThrow(() -> new VendorNotFoundException(id));
        return VendorResponse.from(vendor);
    }

    @Transactional(readOnly = true)
    public List<VendorResponse> listVendors(VendorStatus status) {
        List<Vendor> vendors;
        if (status != null) {
            vendors = vendorRepository.findByStatus(status);
        } else {
            vendors = vendorRepository.findAll();
        }
        return vendors.stream()
                .map(VendorResponse::from)
                .collect(Collectors.toList());
    }

    @Transactional
    public VendorResponse updateVendor(UUID id, UpdateVendorRequest request) {
        Vendor vendor = vendorRepository.findById(id)
                .orElseThrow(() -> new VendorNotFoundException(id));
        if (request.phone() != null) {
            vendor.setPhone(request.phone());
        }
        if (request.website() != null) {
            vendor.setWebsite(request.website());
        }
        if (request.address() != null) {
            vendor.setAddress(request.address());
        }
        if (request.country() != null) {
            vendor.setCountry(request.country());
        }
        if (request.taxId() != null) {
            vendor.setTaxId(request.taxId());
        }
        Vendor saved = vendorRepository.save(vendor);
        return VendorResponse.from(saved);
    }

    @Transactional
    public void suspendVendor(UUID id) {
        Vendor vendor = vendorRepository.findById(id)
                .orElseThrow(() -> new VendorNotFoundException(id));
        vendor.setStatus(VendorStatus.SUSPENDED);
        vendorRepository.save(vendor);
    }

    @Transactional
    public void activateVendor(UUID id) {
        Vendor vendor = vendorRepository.findById(id)
                .orElseThrow(() -> new VendorNotFoundException(id));
        vendor.setStatus(VendorStatus.ACTIVE);
        vendorRepository.save(vendor);
    }

    @Transactional
    public void deleteVendor(UUID id) {
        if (!vendorRepository.existsById(id)) {
            throw new VendorNotFoundException(id);
        }
        vendorRepository.deleteById(id);
    }
}
