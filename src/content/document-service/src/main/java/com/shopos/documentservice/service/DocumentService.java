package com.shopos.documentservice.service;

import com.shopos.documentservice.domain.Document;
import com.shopos.documentservice.domain.DocumentType;
import com.shopos.documentservice.dto.DocumentResponse;
import com.shopos.documentservice.exception.DocumentNotFoundException;
import com.shopos.documentservice.repository.DocumentRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.io.InputStream;
import java.util.List;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class DocumentService {

    private final DocumentRepository documentRepository;
    private final MinioService minioService;

    @Value("${minio.presigned-url-expiry-minutes:60}")
    private int presignedUrlExpiryMinutes;

    @Transactional
    public DocumentResponse uploadDocument(
            MultipartFile file,
            UUID ownerId,
            String description,
            String tags,
            DocumentType documentType
    ) {
        String originalName = file.getOriginalFilename() != null ? file.getOriginalFilename() : file.getName();
        String storedKey = buildStoredKey(ownerId, originalName);

        try (InputStream inputStream = file.getInputStream()) {
            minioService.uploadObject(storedKey, inputStream, file.getContentType(), file.getSize());
        } catch (IOException e) {
            throw new RuntimeException("Failed to read uploaded file", e);
        }

        Document document = Document.builder()
                .ownerId(ownerId)
                .name(sanitizeName(originalName))
                .originalName(originalName)
                .contentType(file.getContentType() != null ? file.getContentType() : "application/octet-stream")
                .size(file.getSize())
                .documentType(documentType != null ? documentType : inferDocumentType(file.getContentType()))
                .storedKey(storedKey)
                .tags(tags)
                .description(description)
                .version(1)
                .build();

        Document saved = documentRepository.save(document);
        log.info("Uploaded document id={} owner={} key={}", saved.getId(), ownerId, storedKey);

        String presignedUrl = minioService.getPresignedUrl(storedKey, presignedUrlExpiryMinutes);
        return DocumentResponse.from(saved, presignedUrl);
    }

    @Transactional(readOnly = true)
    public DocumentResponse getDocument(UUID id) {
        Document document = findOrThrow(id);
        String presignedUrl = minioService.getPresignedUrl(document.getStoredKey(), presignedUrlExpiryMinutes);
        return DocumentResponse.from(document, presignedUrl);
    }

    @Transactional(readOnly = true)
    public List<DocumentResponse> listDocuments(UUID ownerId, DocumentType type) {
        List<Document> docs;
        if (ownerId != null && type != null) {
            docs = documentRepository.findByOwnerIdAndDocumentType(ownerId, type);
        } else if (ownerId != null) {
            docs = documentRepository.findByOwnerId(ownerId);
        } else if (type != null) {
            docs = documentRepository.findByDocumentType(type);
        } else {
            docs = documentRepository.findAll();
        }
        return docs.stream()
                .map(DocumentResponse::from)
                .toList();
    }

    @Transactional(readOnly = true)
    public List<DocumentResponse> searchDocuments(String query) {
        return documentRepository.findByNameContainingIgnoreCase(query)
                .stream()
                .map(DocumentResponse::from)
                .toList();
    }

    @Transactional
    public void deleteDocument(UUID id) {
        Document document = findOrThrow(id);
        minioService.deleteObject(document.getStoredKey());
        documentRepository.delete(document);
        log.info("Deleted document id={}", id);
    }

    @Transactional(readOnly = true)
    public InputStream getDownloadStream(UUID id) {
        Document document = findOrThrow(id);
        return minioService.getObject(document.getStoredKey());
    }

    @Transactional(readOnly = true)
    public Document getRawDocument(UUID id) {
        return findOrThrow(id);
    }

    private Document findOrThrow(UUID id) {
        return documentRepository.findById(id)
                .orElseThrow(() -> new DocumentNotFoundException("Document not found: " + id));
    }

    private String buildStoredKey(UUID ownerId, String originalName) {
        return String.format("documents/%s/%s/%s",
                ownerId,
                UUID.randomUUID(),
                sanitizeName(originalName));
    }

    private String sanitizeName(String name) {
        if (name == null) return "unnamed";
        return name.replaceAll("[^a-zA-Z0-9._\\-]", "_");
    }

    private DocumentType inferDocumentType(String contentType) {
        if (contentType == null) return DocumentType.OTHER;
        return switch (contentType) {
            case "application/pdf" -> DocumentType.PDF;
            case "application/msword",
                 "application/vnd.openxmlformats-officedocument.wordprocessingml.document" -> DocumentType.WORD;
            case "application/vnd.ms-excel",
                 "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" -> DocumentType.EXCEL;
            case "text/plain" -> DocumentType.TEXT;
            default -> contentType.startsWith("image/") ? DocumentType.IMAGE : DocumentType.OTHER;
        };
    }
}
