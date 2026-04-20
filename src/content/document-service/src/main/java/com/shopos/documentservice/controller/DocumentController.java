package com.shopos.documentservice.controller;

import com.shopos.documentservice.domain.Document;
import com.shopos.documentservice.domain.DocumentType;
import com.shopos.documentservice.dto.DocumentResponse;
import com.shopos.documentservice.service.DocumentService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.core.io.InputStreamResource;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.InputStream;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@Slf4j
@RestController
@RequestMapping("/documents")
@RequiredArgsConstructor
public class DocumentController {

    private final DocumentService documentService;

    @PostMapping(consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    public ResponseEntity<DocumentResponse> uploadDocument(
            @RequestParam("file") MultipartFile file,
            @RequestParam("ownerId") UUID ownerId,
            @RequestParam(value = "description", required = false) String description,
            @RequestParam(value = "tags", required = false) String tags,
            @RequestParam(value = "documentType", required = false) DocumentType documentType
    ) {
        DocumentResponse response = documentService.uploadDocument(file, ownerId, description, tags, documentType);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/{id}")
    public ResponseEntity<DocumentResponse> getDocument(@PathVariable UUID id) {
        return ResponseEntity.ok(documentService.getDocument(id));
    }

    @GetMapping
    public ResponseEntity<List<DocumentResponse>> listDocuments(
            @RequestParam(required = false) UUID ownerId,
            @RequestParam(required = false) DocumentType type
    ) {
        return ResponseEntity.ok(documentService.listDocuments(ownerId, type));
    }

    @GetMapping("/search")
    public ResponseEntity<List<DocumentResponse>> searchDocuments(@RequestParam("q") String query) {
        if (query == null || query.isBlank()) {
            throw new IllegalArgumentException("Search query must not be blank");
        }
        return ResponseEntity.ok(documentService.searchDocuments(query));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteDocument(@PathVariable UUID id) {
        documentService.deleteDocument(id);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/{id}/download")
    public ResponseEntity<InputStreamResource> downloadDocument(@PathVariable UUID id) {
        Document doc = documentService.getRawDocument(id);
        InputStream stream = documentService.getDownloadStream(id);

        return ResponseEntity.ok()
                .header(HttpHeaders.CONTENT_DISPOSITION,
                        "attachment; filename=\"" + doc.getOriginalName() + "\"")
                .contentType(MediaType.parseMediaType(doc.getContentType()))
                .contentLength(doc.getSize())
                .body(new InputStreamResource(stream));
    }

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
