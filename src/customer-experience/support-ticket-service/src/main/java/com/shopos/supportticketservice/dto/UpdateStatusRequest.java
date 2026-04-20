package com.shopos.supportticketservice.dto;

import com.shopos.supportticketservice.domain.TicketStatus;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class UpdateStatusRequest {

    @NotNull(message = "status is required")
    private TicketStatus status;
}
