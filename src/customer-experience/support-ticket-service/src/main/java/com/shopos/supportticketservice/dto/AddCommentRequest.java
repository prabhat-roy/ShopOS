package com.shopos.supportticketservice.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AddCommentRequest {

    @NotBlank(message = "authorId is required")
    private String authorId;

    @NotBlank(message = "authorType is required")
    @Pattern(regexp = "CUSTOMER|AGENT", message = "authorType must be CUSTOMER or AGENT")
    private String authorType;

    @NotBlank(message = "body is required")
    private String body;

    private boolean internal = false;
}
