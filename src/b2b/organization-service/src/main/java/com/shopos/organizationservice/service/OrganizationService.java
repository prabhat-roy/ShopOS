package com.shopos.organizationservice.service;

import com.shopos.organizationservice.domain.OrgMember;
import com.shopos.organizationservice.domain.OrgStatus;
import com.shopos.organizationservice.domain.OrgType;
import com.shopos.organizationservice.domain.Organization;
import com.shopos.organizationservice.dto.CreateOrgRequest;
import com.shopos.organizationservice.dto.InviteMemberRequest;
import com.shopos.organizationservice.dto.MemberResponse;
import com.shopos.organizationservice.dto.OrgResponse;
import com.shopos.organizationservice.dto.UpdateOrgRequest;
import com.shopos.organizationservice.exception.ConflictException;
import com.shopos.organizationservice.exception.NotFoundException;
import com.shopos.organizationservice.repository.OrgMemberRepository;
import com.shopos.organizationservice.repository.OrganizationRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class OrganizationService {

    private final OrganizationRepository organizationRepository;
    private final OrgMemberRepository orgMemberRepository;

    @Transactional
    public OrgResponse createOrg(CreateOrgRequest request) {
        if (organizationRepository.existsByEmail(request.email())) {
            throw new ConflictException("An organization with email '" + request.email() + "' already exists");
        }

        String baseSlug = request.name().toLowerCase().replaceAll("[^a-z0-9]+", "-");
        String shortUuid = UUID.randomUUID().toString().replace("-", "").substring(0, 8);
        String slug = baseSlug + "-" + shortUuid;

        Organization org = Organization.builder()
                .name(request.name())
                .slug(slug)
                .email(request.email())
                .phone(request.phone())
                .website(request.website())
                .type(request.type() != null ? request.type() : OrgType.SMB)
                .status(OrgStatus.PENDING_VERIFICATION)
                .industry(request.industry())
                .country(request.country())
                .address(request.address())
                .employeeCount(request.employeeCount() != null ? request.employeeCount() : 0)
                .build();

        return OrgResponse.from(organizationRepository.save(org));
    }

    public OrgResponse getOrg(UUID id) {
        Organization org = organizationRepository.findById(id)
                .orElseThrow(() -> new NotFoundException("Organization", id));
        return OrgResponse.from(org);
    }

    public OrgResponse getBySlug(String slug) {
        Organization org = organizationRepository.findBySlug(slug)
                .orElseThrow(() -> new NotFoundException("Organization not found with slug: " + slug));
        return OrgResponse.from(org);
    }

    public List<OrgResponse> listOrgs(OrgStatus status, OrgType type) {
        List<Organization> orgs;
        if (status != null && type != null) {
            orgs = organizationRepository.findByStatusAndType(status, type);
        } else if (status != null) {
            orgs = organizationRepository.findByStatus(status);
        } else if (type != null) {
            orgs = organizationRepository.findByType(type);
        } else {
            orgs = organizationRepository.findAll();
        }
        return orgs.stream().map(OrgResponse::from).collect(Collectors.toList());
    }

    @Transactional
    public OrgResponse updateOrg(UUID id, UpdateOrgRequest request) {
        Organization org = organizationRepository.findById(id)
                .orElseThrow(() -> new NotFoundException("Organization", id));

        if (request.name() != null) org.setName(request.name());
        if (request.phone() != null) org.setPhone(request.phone());
        if (request.website() != null) org.setWebsite(request.website());
        if (request.type() != null) org.setType(request.type());
        if (request.industry() != null) org.setIndustry(request.industry());
        if (request.taxId() != null) org.setTaxId(request.taxId());
        if (request.country() != null) org.setCountry(request.country());
        if (request.address() != null) org.setAddress(request.address());
        if (request.employeeCount() != null) org.setEmployeeCount(request.employeeCount());
        if (request.creditLimit() != null) org.setCreditLimit(request.creditLimit());
        if (request.parentOrgId() != null) org.setParentOrgId(request.parentOrgId());
        if (request.settings() != null) org.setSettings(request.settings());

        return OrgResponse.from(organizationRepository.save(org));
    }

    @Transactional
    public void suspendOrg(UUID id) {
        Organization org = organizationRepository.findById(id)
                .orElseThrow(() -> new NotFoundException("Organization", id));
        if (org.getStatus() == OrgStatus.SUSPENDED) {
            throw new IllegalStateException("Organization is already suspended");
        }
        org.setStatus(OrgStatus.SUSPENDED);
        organizationRepository.save(org);
    }

    @Transactional
    public void activateOrg(UUID id) {
        Organization org = organizationRepository.findById(id)
                .orElseThrow(() -> new NotFoundException("Organization", id));
        org.setStatus(OrgStatus.ACTIVE);
        organizationRepository.save(org);
    }

    public List<OrgResponse> getSubsidiaries(UUID parentId) {
        organizationRepository.findById(parentId)
                .orElseThrow(() -> new NotFoundException("Organization", parentId));
        return organizationRepository.findByParentOrgId(parentId)
                .stream()
                .map(OrgResponse::from)
                .collect(Collectors.toList());
    }

    @Transactional
    public MemberResponse inviteMember(UUID orgId, InviteMemberRequest request) {
        organizationRepository.findById(orgId)
                .orElseThrow(() -> new NotFoundException("Organization", orgId));

        orgMemberRepository.findByOrgIdAndUserId(orgId, request.userId()).ifPresent(existing -> {
            if (existing.isActive()) {
                throw new ConflictException("User " + request.userId() + " is already an active member of this organization");
            }
        });

        OrgMember member = OrgMember.builder()
                .orgId(orgId)
                .userId(request.userId())
                .role(request.role())
                .department(request.department())
                .jobTitle(request.jobTitle())
                .active(true)
                .invitedAt(LocalDateTime.now())
                .build();

        return MemberResponse.from(orgMemberRepository.save(member));
    }

    public MemberResponse getMember(UUID orgId, UUID memberId) {
        OrgMember member = orgMemberRepository.findById(memberId)
                .filter(m -> m.getOrgId().equals(orgId))
                .orElseThrow(() -> new NotFoundException("OrgMember", memberId));
        return MemberResponse.from(member);
    }

    public List<MemberResponse> listMembers(UUID orgId, boolean activeOnly) {
        organizationRepository.findById(orgId)
                .orElseThrow(() -> new NotFoundException("Organization", orgId));

        List<OrgMember> members = activeOnly
                ? orgMemberRepository.findByOrgIdAndActiveTrue(orgId)
                : orgMemberRepository.findByOrgId(orgId);

        return members.stream().map(MemberResponse::from).collect(Collectors.toList());
    }

    @Transactional
    public void updateMemberRole(UUID orgId, UUID memberId, String role) {
        OrgMember member = orgMemberRepository.findById(memberId)
                .filter(m -> m.getOrgId().equals(orgId))
                .orElseThrow(() -> new NotFoundException("OrgMember", memberId));
        member.setRole(role);
        orgMemberRepository.save(member);
    }

    @Transactional
    public void removeMember(UUID orgId, UUID memberId) {
        OrgMember member = orgMemberRepository.findById(memberId)
                .filter(m -> m.getOrgId().equals(orgId))
                .orElseThrow(() -> new NotFoundException("OrgMember", memberId));
        member.setActive(false);
        orgMemberRepository.save(member);
    }

    public List<MemberResponse> getOrgsByUserId(UUID userId) {
        return orgMemberRepository.findByUserId(userId)
                .stream()
                .map(MemberResponse::from)
                .collect(Collectors.toList());
    }
}
