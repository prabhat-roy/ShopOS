package com.shopos.documentservice.service;

import com.shopos.documentservice.domain.Document;
import com.shopos.documentservice.domain.DocumentType;
import com.shopos.documentservice.dto.DocumentResponse;
import com.shopos.documentservice.exception.DocumentNotFoundException;
import com.shopos.documentservice.repository.DocumentRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.mock.web.MockMultipartFile;
import org.springframework.test.util.ReflectionTestUtils;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class DocumentServiceTest {

    @Mock
    private DocumentRepository documentRepository;

    @Mock
    private MinioService minioService;

    @InjectMocks
    private DocumentService documentService;

    private UUID ownerId;
    private UUID docId;
    private Document sampleDocument;

    @BeforeEach
    void setUp() {
        ReflectionTestUtils.setField(documentService, "presignedUrlExpiryMinutes", 60);
        ownerId = UUID.randomUUID();
        docId = UUID.randomUUID();
        sampleDocument = Document.builder()
                .id(docId)
                .ownerId(ownerId)
                .name("report.pdf")
                .originalName("report.pdf")
                .contentType("application/pdf")
                .size(1024L)
                .documentType(DocumentType.PDF)
                .storedKey("documents/" + ownerId + "/uuid/report.pdf")
                .description("Annual report")
                .tags("finance,annual")
                .version(1)
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .build();
    }

    @Test
    @DisplayName("uploadDocument — saves metadata and uploads to MinIO")
    void uploadDocument_savesMetadataAndUploadsToMinio() {
        MockMultipartFile file = new MockMultipartFile(
                "file", "report.pdf", "application/pdf", "PDF content".getBytes()
        );
        when(documentRepository.save(any(Document.class))).thenReturn(sampleDocument);
        when(minioService.getPresignedUrl(anyString(), anyInt())).thenReturn("https://minio/presigned");

        DocumentResponse result = documentService.uploadDocument(
                file, ownerId, "Annual report", "finance,annual", DocumentType.PDF
        );

        verify(minioService).uploadObject(anyString(), any(), eq("application/pdf"), anyLong());
        verify(documentRepository).save(any(Document.class));
        assertThat(result.presignedUrl()).isEqualTo("https://minio/presigned");
        assertThat(result.ownerId()).isEqualTo(ownerId);
    }

    @Test
    @DisplayName("uploadDocument — infers document type from content type when not provided")
    void uploadDocument_infersDocumentTypeFromContentType() {
        MockMultipartFile file = new MockMultipartFile(
                "file", "data.xlsx",
                "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
                "data".getBytes()
        );
        ArgumentCaptor<Document> captor = ArgumentCaptor.forClass(Document.class);
        when(documentRepository.save(captor.capture())).thenAnswer(inv -> {
            Document d = inv.getArgument(0);
            d = Document.builder()
                    .id(UUID.randomUUID())
                    .ownerId(d.getOwnerId())
                    .name(d.getName())
                    .originalName(d.getOriginalName())
                    .contentType(d.getContentType())
                    .size(d.getSize())
                    .documentType(d.getDocumentType())
                    .storedKey(d.getStoredKey())
                    .version(1)
                    .createdAt(Instant.now())
                    .updatedAt(Instant.now())
                    .build();
            return d;
        });
        when(minioService.getPresignedUrl(anyString(), anyInt())).thenReturn("http://url");

        documentService.uploadDocument(file, ownerId, null, null, null);

        assertThat(captor.getValue().getDocumentType()).isEqualTo(DocumentType.EXCEL);
    }

    @Test
    @DisplayName("getDocument — returns document with presigned URL")
    void getDocument_returnsWithPresignedUrl() {
        when(documentRepository.findById(docId)).thenReturn(Optional.of(sampleDocument));
        when(minioService.getPresignedUrl(sampleDocument.getStoredKey(), 60))
                .thenReturn("https://minio/presigned");

        DocumentResponse result = documentService.getDocument(docId);

        assertThat(result.id()).isEqualTo(docId);
        assertThat(result.presignedUrl()).isEqualTo("https://minio/presigned");
    }

    @Test
    @DisplayName("getDocument — throws DocumentNotFoundException when not found")
    void getDocument_throwsWhenNotFound() {
        when(documentRepository.findById(docId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> documentService.getDocument(docId))
                .isInstanceOf(DocumentNotFoundException.class)
                .hasMessageContaining(docId.toString());
    }

    @Test
    @DisplayName("listDocuments — filters by ownerId")
    void listDocuments_filtersByOwnerId() {
        when(documentRepository.findByOwnerId(ownerId)).thenReturn(List.of(sampleDocument));

        List<DocumentResponse> results = documentService.listDocuments(ownerId, null);

        assertThat(results).hasSize(1);
        assertThat(results.get(0).ownerId()).isEqualTo(ownerId);
        verify(documentRepository).findByOwnerId(ownerId);
    }

    @Test
    @DisplayName("listDocuments — filters by documentType")
    void listDocuments_filtersByDocumentType() {
        when(documentRepository.findByDocumentType(DocumentType.PDF))
                .thenReturn(List.of(sampleDocument));

        List<DocumentResponse> results = documentService.listDocuments(null, DocumentType.PDF);

        assertThat(results).hasSize(1);
        assertThat(results.get(0).documentType()).isEqualTo(DocumentType.PDF);
        verify(documentRepository).findByDocumentType(DocumentType.PDF);
    }

    @Test
    @DisplayName("searchDocuments — returns matching documents")
    void searchDocuments_returnsMatches() {
        when(documentRepository.findByNameContainingIgnoreCase("report"))
                .thenReturn(List.of(sampleDocument));

        List<DocumentResponse> results = documentService.searchDocuments("report");

        assertThat(results).hasSize(1);
        assertThat(results.get(0).name()).isEqualTo("report.pdf");
    }

    @Test
    @DisplayName("deleteDocument — removes from MinIO and DB")
    void deleteDocument_removesFromMinioAndDb() {
        when(documentRepository.findById(docId)).thenReturn(Optional.of(sampleDocument));

        documentService.deleteDocument(docId);

        verify(minioService).deleteObject(sampleDocument.getStoredKey());
        verify(documentRepository).delete(sampleDocument);
    }

    @Test
    @DisplayName("deleteDocument — throws when document does not exist")
    void deleteDocument_throwsWhenNotFound() {
        when(documentRepository.findById(docId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> documentService.deleteDocument(docId))
                .isInstanceOf(DocumentNotFoundException.class);

        verifyNoInteractions(minioService);
    }
}
