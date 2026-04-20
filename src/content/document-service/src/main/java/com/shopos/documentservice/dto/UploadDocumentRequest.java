package com.shopos.documentservice.dto;

import com.shopos.documentservice.domain.DocumentType;

import java.util.UUID;

public record UploadDocumentRequest(
        UUID ownerId,
        String description,
        String tags,
        DocumentType documentType
) {
}
