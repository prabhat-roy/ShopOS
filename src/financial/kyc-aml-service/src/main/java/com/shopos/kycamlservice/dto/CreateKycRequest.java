package com.shopos.kycamlservice.dto;

import jakarta.validation.constraints.*;

import java.time.LocalDate;
import java.util.UUID;

public record CreateKycRequest(

    @NotNull(message = "customerId is required")
    UUID customerId,

    @NotBlank(message = "firstName is required")
    @Size(max = 100, message = "firstName must not exceed 100 characters")
    String firstName,

    @NotBlank(message = "lastName is required")
    @Size(max = 100, message = "lastName must not exceed 100 characters")
    String lastName,

    @NotNull(message = "dateOfBirth is required")
    @Past(message = "dateOfBirth must be a past date")
    LocalDate dateOfBirth,

    @NotBlank(message = "nationality is required")
    @Size(min = 2, max = 2, message = "nationality must be a 2-letter ISO country code")
    String nationality,

    @NotBlank(message = "documentType is required")
    @Pattern(
        regexp = "PASSPORT|NATIONAL_ID|DRIVERS_LICENSE",
        message = "documentType must be one of: PASSPORT, NATIONAL_ID, DRIVERS_LICENSE"
    )
    String documentType,

    @NotBlank(message = "documentNumber is required")
    @Size(max = 100, message = "documentNumber must not exceed 100 characters")
    String documentNumber,

    @NotNull(message = "documentExpiry is required")
    @Future(message = "documentExpiry must be a future date")
    LocalDate documentExpiry,

    String notes
) {}
