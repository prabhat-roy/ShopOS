package com.shopos.documentservice.dto;

import com.shopos.documentservice.domain.Document;
import com.shopos.documentservice.domain.DocumentType;

import java.time.Instant;
import java.util.UUID;

public record DocumentResponse(
        UUID id,
        UUID ownerId,
        String name,
        String originalName,
        String contentType,
        long size,
        DocumentType documentType,
        String storedKey,
        String tags,
        String description,
        int version,
        String presignedUrl,
        Instant createdAt,
        Instant updatedAt
) {

    public static DocumentResponse from(Document document, String presignedUrl) {
        return new DocumentResponse(
                document.getId(),
                document.getOwnerId(),
                document.getName(),
                document.getOriginalName(),
                document.getContentType(),
                document.getSize(),
                document.getDocumentType(),
                document.getStoredKey(),
                document.getTags(),
                document.getDescription(),
                document.getVersion(),
                presignedUrl,
                document.getCreatedAt(),
                document.getUpdatedAt()
        );
    }

    public static DocumentResponse from(Document document) {
        return from(document, null);
    }
}
