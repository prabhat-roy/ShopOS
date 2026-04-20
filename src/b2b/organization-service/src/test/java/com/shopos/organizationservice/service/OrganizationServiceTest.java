package com.shopos.organizationservice.service;

import com.shopos.organizationservice.domain.OrgMember;
import com.shopos.organizationservice.domain.OrgStatus;
import com.shopos.organizationservice.domain.OrgType;
import com.shopos.organizationservice.domain.Organization;
import com.shopos.organizationservice.dto.CreateOrgRequest;
import com.shopos.organizationservice.dto.InviteMemberRequest;
import com.shopos.organizationservice.dto.MemberResponse;
import com.shopos.organizationservice.dto.OrgResponse;
import com.shopos.organizationservice.exception.ConflictException;
import com.shopos.organizationservice.exception.NotFoundException;
import com.shopos.organizationservice.repository.OrgMemberRepository;
import com.shopos.organizationservice.repository.OrganizationRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrganizationServiceTest {

    @Mock
    private OrganizationRepository organizationRepository;

    @Mock
    private OrgMemberRepository orgMemberRepository;

    @InjectMocks
    private OrganizationService organizationService;

    // ── helpers ──────────────────────────────────────────────────────────────

    private Organization buildOrg(UUID id, OrgStatus status) {
        return Organization.builder()
                .id(id)
                .name("Acme Corp")
                .slug("acme-corp-abcd1234")
                .email("admin@acme.com")
                .type(OrgType.ENTERPRISE)
                .status(status)
                .employeeCount(500)
                .creditLimit(BigDecimal.valueOf(100_000))
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();
    }

    private OrgMember buildMember(UUID id, UUID orgId, UUID userId) {
        return OrgMember.builder()
                .id(id)
                .orgId(orgId)
                .userId(userId)
                .role("ADMIN")
                .active(true)
                .invitedAt(LocalDateTime.now())
                .createdAt(LocalDateTime.now())
                .build();
    }

    // ── Test 1: createOrg success ─────────────────────────────────────────────

    @Test
    @DisplayName("createOrg: should persist org with PENDING_VERIFICATION status and generated slug")
    void createOrg_success() {
        CreateOrgRequest req = new CreateOrgRequest(
                "Acme Corp", "admin@acme.com", null, null, OrgType.ENTERPRISE,
                "Technology", null, "US", null, 500);

        when(organizationRepository.existsByEmail("admin@acme.com")).thenReturn(false);

        ArgumentCaptor<Organization> captor = ArgumentCaptor.forClass(Organization.class);
        UUID generatedId = UUID.randomUUID();
        Organization saved = buildOrg(generatedId, OrgStatus.PENDING_VERIFICATION);
        when(organizationRepository.save(captor.capture())).thenReturn(saved);

        OrgResponse response = organizationService.createOrg(req);

        assertThat(response).isNotNull();
        assertThat(response.status()).isEqualTo(OrgStatus.PENDING_VERIFICATION);
        Organization captured = captor.getValue();
        assertThat(captured.getSlug()).matches("acme-corp-[a-z0-9]{8}");
        assertThat(captured.getEmail()).isEqualTo("admin@acme.com");
        verify(organizationRepository).existsByEmail("admin@acme.com");
        verify(organizationRepository).save(any(Organization.class));
    }

    // ── Test 2: createOrg duplicate email ────────────────────────────────────

    @Test
    @DisplayName("createOrg: should throw ConflictException when email already exists")
    void createOrg_duplicateEmail_throwsConflict() {
        CreateOrgRequest req = new CreateOrgRequest(
                "Acme Corp", "admin@acme.com", null, null, null,
                null, null, null, null, null);

        when(organizationRepository.existsByEmail("admin@acme.com")).thenReturn(true);

        assertThatThrownBy(() -> organizationService.createOrg(req))
                .isInstanceOf(ConflictException.class)
                .hasMessageContaining("admin@acme.com");

        verify(organizationRepository, never()).save(any());
    }

    // ── Test 3: getOrg found ──────────────────────────────────────────────────

    @Test
    @DisplayName("getOrg: should return OrgResponse when organization exists")
    void getOrg_found() {
        UUID id = UUID.randomUUID();
        Organization org = buildOrg(id, OrgStatus.ACTIVE);
        when(organizationRepository.findById(id)).thenReturn(Optional.of(org));

        OrgResponse response = organizationService.getOrg(id);

        assertThat(response.id()).isEqualTo(id);
        assertThat(response.status()).isEqualTo(OrgStatus.ACTIVE);
    }

    // ── Test 4: getOrg not found ──────────────────────────────────────────────

    @Test
    @DisplayName("getOrg: should throw NotFoundException when organization does not exist")
    void getOrg_notFound_throwsNotFoundException() {
        UUID id = UUID.randomUUID();
        when(organizationRepository.findById(id)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> organizationService.getOrg(id))
                .isInstanceOf(NotFoundException.class)
                .hasMessageContaining(id.toString());
    }

    // ── Test 5: listOrgs filtered by status ──────────────────────────────────

    @Test
    @DisplayName("listOrgs: should return only orgs matching status filter")
    void listOrgs_filteredByStatus() {
        UUID id1 = UUID.randomUUID();
        UUID id2 = UUID.randomUUID();
        List<Organization> activeOrgs = List.of(
                buildOrg(id1, OrgStatus.ACTIVE),
                buildOrg(id2, OrgStatus.ACTIVE));

        when(organizationRepository.findByStatus(OrgStatus.ACTIVE)).thenReturn(activeOrgs);

        List<OrgResponse> result = organizationService.listOrgs(OrgStatus.ACTIVE, null);

        assertThat(result).hasSize(2);
        assertThat(result).allMatch(r -> r.status() == OrgStatus.ACTIVE);
        verify(organizationRepository).findByStatus(OrgStatus.ACTIVE);
        verify(organizationRepository, never()).findAll();
    }

    // ── Test 6: updateOrg ────────────────────────────────────────────────────

    @Test
    @DisplayName("updateOrg: should update mutable fields and save")
    void updateOrg_updatesFields() {
        UUID id = UUID.randomUUID();
        Organization org = buildOrg(id, OrgStatus.ACTIVE);
        when(organizationRepository.findById(id)).thenReturn(Optional.of(org));
        when(organizationRepository.save(org)).thenReturn(org);

        com.shopos.organizationservice.dto.UpdateOrgRequest req =
                new com.shopos.organizationservice.dto.UpdateOrgRequest(
                        "Acme Corp Updated", "+1-555-0001", "https://acme.com",
                        OrgType.SMB, "Retail", "US-TAX-123", "CA",
                        "123 Main St", 600, BigDecimal.valueOf(200_000), null, null);

        OrgResponse response = organizationService.updateOrg(id, req);

        assertThat(org.getName()).isEqualTo("Acme Corp Updated");
        assertThat(org.getEmployeeCount()).isEqualTo(600);
        verify(organizationRepository).save(org);
    }

    // ── Test 7: suspendOrg ───────────────────────────────────────────────────

    @Test
    @DisplayName("suspendOrg: should set status to SUSPENDED")
    void suspendOrg_setsStatusSuspended() {
        UUID id = UUID.randomUUID();
        Organization org = buildOrg(id, OrgStatus.ACTIVE);
        when(organizationRepository.findById(id)).thenReturn(Optional.of(org));
        when(organizationRepository.save(org)).thenReturn(org);

        organizationService.suspendOrg(id);

        assertThat(org.getStatus()).isEqualTo(OrgStatus.SUSPENDED);
        verify(organizationRepository).save(org);
    }

    // ── Test 8: activateOrg ──────────────────────────────────────────────────

    @Test
    @DisplayName("activateOrg: should set status to ACTIVE")
    void activateOrg_setsStatusActive() {
        UUID id = UUID.randomUUID();
        Organization org = buildOrg(id, OrgStatus.PENDING_VERIFICATION);
        when(organizationRepository.findById(id)).thenReturn(Optional.of(org));
        when(organizationRepository.save(org)).thenReturn(org);

        organizationService.activateOrg(id);

        assertThat(org.getStatus()).isEqualTo(OrgStatus.ACTIVE);
        verify(organizationRepository).save(org);
    }

    // ── Test 9: inviteMember ─────────────────────────────────────────────────

    @Test
    @DisplayName("inviteMember: should save new member with active=true")
    void inviteMember_savesNewMember() {
        UUID orgId = UUID.randomUUID();
        UUID userId = UUID.randomUUID();
        Organization org = buildOrg(orgId, OrgStatus.ACTIVE);

        when(organizationRepository.findById(orgId)).thenReturn(Optional.of(org));
        when(orgMemberRepository.findByOrgIdAndUserId(orgId, userId)).thenReturn(Optional.empty());

        OrgMember savedMember = buildMember(UUID.randomUUID(), orgId, userId);
        when(orgMemberRepository.save(any(OrgMember.class))).thenReturn(savedMember);

        InviteMemberRequest req = new InviteMemberRequest(userId, "ADMIN", "Engineering", "Senior Engineer");
        MemberResponse response = organizationService.inviteMember(orgId, req);

        assertThat(response.orgId()).isEqualTo(orgId);
        assertThat(response.userId()).isEqualTo(userId);
        assertThat(response.active()).isTrue();
        verify(orgMemberRepository).save(any(OrgMember.class));
    }

    // ── Test 10: listMembers activeOnly ──────────────────────────────────────

    @Test
    @DisplayName("listMembers: should return only active members when activeOnly=true")
    void listMembers_activeOnly() {
        UUID orgId = UUID.randomUUID();
        Organization org = buildOrg(orgId, OrgStatus.ACTIVE);
        UUID userId1 = UUID.randomUUID();
        UUID userId2 = UUID.randomUUID();

        List<OrgMember> activeMembers = List.of(
                buildMember(UUID.randomUUID(), orgId, userId1),
                buildMember(UUID.randomUUID(), orgId, userId2));

        when(organizationRepository.findById(orgId)).thenReturn(Optional.of(org));
        when(orgMemberRepository.findByOrgIdAndActiveTrue(orgId)).thenReturn(activeMembers);

        List<MemberResponse> result = organizationService.listMembers(orgId, true);

        assertThat(result).hasSize(2);
        assertThat(result).allMatch(MemberResponse::active);
        verify(orgMemberRepository).findByOrgIdAndActiveTrue(orgId);
        verify(orgMemberRepository, never()).findByOrgId(any());
    }
}
