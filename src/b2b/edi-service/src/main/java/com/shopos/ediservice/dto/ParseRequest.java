package com.shopos.ediservice.dto;

import com.shopos.ediservice.domain.EdiFormat;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

/**
 * Request payload for the EDI parse endpoint.
 *
 * @param content Raw EDI message string (X12 or EDIFACT).
 * @param format  EDI format standard to use when parsing.
 */
public record ParseRequest(
        @NotBlank(message = "EDI content must not be blank")
        String content,

        @NotNull(message = "EDI format must be specified")
        EdiFormat format
) {}
