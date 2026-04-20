package com.shopos.supportticketservice.dto;

import com.shopos.supportticketservice.domain.TicketCategory;
import com.shopos.supportticketservice.domain.TicketPriority;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateTicketRequest {

    @NotNull(message = "customerId is required")
    private UUID customerId;

    private UUID orderId;

    @NotBlank(message = "subject is required")
    @Size(max = 255, message = "subject must be 255 characters or fewer")
    private String subject;

    @NotBlank(message = "description is required")
    private String description;

    @NotNull(message = "category is required")
    private TicketCategory category;

    private TicketPriority priority;
}
