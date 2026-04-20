package com.shopos.adservice.dto;

import com.shopos.adservice.domain.AdType;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.util.List;

public record ServeAdRequest(
        @NotBlank String userId,
        @NotBlank String sessionId,
        @NotBlank String placementId,
        List<String> categories,
        @NotNull AdType adType
) {}
