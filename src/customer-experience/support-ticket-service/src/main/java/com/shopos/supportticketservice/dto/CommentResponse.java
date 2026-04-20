package com.shopos.supportticketservice.dto;

import com.shopos.supportticketservice.domain.TicketComment;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CommentResponse {

    private UUID id;
    private UUID ticketId;
    private String authorId;
    private String authorType;
    private String body;
    private boolean internal;
    private LocalDateTime createdAt;

    public static CommentResponse from(TicketComment comment) {
        return CommentResponse.builder()
                .id(comment.getId())
                .ticketId(comment.getTicketId())
                .authorId(comment.getAuthorId())
                .authorType(comment.getAuthorType())
                .body(comment.getBody())
                .internal(comment.isInternal())
                .createdAt(comment.getCreatedAt())
                .build();
    }
}
