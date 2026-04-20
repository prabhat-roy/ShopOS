package com.shopos.promotionsservice.dto;

import java.math.BigDecimal;

public record ValidateResponse(
        boolean valid,
        String code,
        BigDecimal discountAmount,
        String reason
) {

    /** Convenience factory for a successful validation. */
    public static ValidateResponse ok(String code, BigDecimal discountAmount) {
        return new ValidateResponse(true, code, discountAmount, null);
    }

    /** Convenience factory for a failed validation. */
    public static ValidateResponse fail(String code, String reason) {
        return new ValidateResponse(false, code, BigDecimal.ZERO, reason);
    }
}
