package com.shopos.vendorservice.dto;

import com.shopos.vendorservice.domain.Vendor;
import com.shopos.vendorservice.domain.VendorStatus;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.UUID;

public record VendorResponse(
        UUID id,
        String name,
        String email,
        String phone,
        String website,
        VendorStatus status,
        String country,
        String address,
        String taxId,
        BigDecimal rating,
        int totalOrders,
        Instant createdAt,
        Instant updatedAt
) {
    public static VendorResponse from(Vendor vendor) {
        return new VendorResponse(
                vendor.getId(),
                vendor.getName(),
                vendor.getEmail(),
                vendor.getPhone(),
                vendor.getWebsite(),
                vendor.getStatus(),
                vendor.getCountry(),
                vendor.getAddress(),
                vendor.getTaxId(),
                vendor.getRating(),
                vendor.getTotalOrders(),
                vendor.getCreatedAt(),
                vendor.getUpdatedAt()
        );
    }
}
