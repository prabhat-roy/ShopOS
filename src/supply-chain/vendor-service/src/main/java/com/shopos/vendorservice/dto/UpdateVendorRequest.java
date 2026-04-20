package com.shopos.vendorservice.dto;

import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

public record UpdateVendorRequest(

        @Pattern(regexp = "^[+]?[0-9\\s\\-().]{7,50}$", message = "Phone number format is invalid")
        String phone,

        @Size(max = 255, message = "Website must not exceed 255 characters")
        String website,

        String address,

        @Size(max = 100, message = "Country must not exceed 100 characters")
        String country,

        @Size(max = 100, message = "Tax ID must not exceed 100 characters")
        String taxId
) {}
