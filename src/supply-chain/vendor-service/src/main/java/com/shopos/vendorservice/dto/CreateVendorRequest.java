package com.shopos.vendorservice.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

public record CreateVendorRequest(

        @NotBlank(message = "Vendor name is required")
        @Size(max = 255, message = "Name must not exceed 255 characters")
        String name,

        @NotBlank(message = "Email is required")
        @Email(message = "Email must be a valid email address")
        @Size(max = 255, message = "Email must not exceed 255 characters")
        String email,

        @NotBlank(message = "Phone is required")
        @Pattern(regexp = "^[+]?[0-9\\s\\-().]{7,50}$", message = "Phone number format is invalid")
        String phone,

        String website,

        String country,

        String address,

        String taxId
) {}
