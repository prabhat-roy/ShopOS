package com.shopos.documentservice.repository;

import com.shopos.documentservice.domain.Document;
import com.shopos.documentservice.domain.DocumentType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface DocumentRepository extends JpaRepository<Document, UUID> {

    List<Document> findByOwnerId(UUID ownerId);

    List<Document> findByDocumentType(DocumentType documentType);

    List<Document> findByOwnerIdAndDocumentType(UUID ownerId, DocumentType documentType);

    List<Document> findByNameContainingIgnoreCase(String name);
}
