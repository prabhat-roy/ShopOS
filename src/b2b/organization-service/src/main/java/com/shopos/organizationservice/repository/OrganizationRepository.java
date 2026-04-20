package com.shopos.organizationservice.repository;

import com.shopos.organizationservice.domain.OrgStatus;
import com.shopos.organizationservice.domain.OrgType;
import com.shopos.organizationservice.domain.Organization;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface OrganizationRepository extends JpaRepository<Organization, UUID> {

    Optional<Organization> findBySlug(String slug);

    List<Organization> findByStatus(OrgStatus status);

    List<Organization> findByType(OrgType type);

    List<Organization> findByStatusAndType(OrgStatus status, OrgType type);

    List<Organization> findByParentOrgId(UUID parentOrgId);

    boolean existsBySlug(String slug);

    boolean existsByEmail(String email);
}
