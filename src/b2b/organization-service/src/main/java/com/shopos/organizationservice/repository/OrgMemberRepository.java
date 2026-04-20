package com.shopos.organizationservice.repository;

import com.shopos.organizationservice.domain.OrgMember;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface OrgMemberRepository extends JpaRepository<OrgMember, UUID> {

    List<OrgMember> findByOrgId(UUID orgId);

    List<OrgMember> findByUserId(UUID userId);

    Optional<OrgMember> findByOrgIdAndUserId(UUID orgId, UUID userId);

    List<OrgMember> findByOrgIdAndActiveTrue(UUID orgId);
}
