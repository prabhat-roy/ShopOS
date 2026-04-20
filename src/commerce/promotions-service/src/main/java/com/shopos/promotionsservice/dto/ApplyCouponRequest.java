package com.shopos.promotionsservice.dto;

import jakarta.validation.constraints.NotBlank;

public record ApplyCouponRequest(

        @NotBlank(message = "code is required")
        String code
) {}
