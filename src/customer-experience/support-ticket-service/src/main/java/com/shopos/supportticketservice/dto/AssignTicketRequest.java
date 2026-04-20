package com.shopos.supportticketservice.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class AssignTicketRequest {

    @NotBlank(message = "agentId is required")
    private String agentId;
}
