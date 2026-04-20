package com.shopos.organizationservice.controller;

import com.shopos.organizationservice.domain.OrgStatus;
import com.shopos.organizationservice.domain.OrgType;
import com.shopos.organizationservice.dto.CreateOrgRequest;
import com.shopos.organizationservice.dto.InviteMemberRequest;
import com.shopos.organizationservice.dto.MemberResponse;
import com.shopos.organizationservice.dto.OrgResponse;
import com.shopos.organizationservice.dto.UpdateMemberRoleRequest;
import com.shopos.organizationservice.dto.UpdateOrgRequest;
import com.shopos.organizationservice.service.OrganizationService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequiredArgsConstructor
public class OrganizationController {

    private final OrganizationService organizationService;

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    // ─── Organization endpoints ───────────────────────────────────────────────

    @PostMapping("/orgs")
    public ResponseEntity<OrgResponse> createOrg(@Valid @RequestBody CreateOrgRequest request) {
        OrgResponse response = organizationService.createOrg(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/orgs/{id}")
    public ResponseEntity<OrgResponse> getOrg(@PathVariable UUID id) {
        return ResponseEntity.ok(organizationService.getOrg(id));
    }

    @GetMapping("/orgs/slug/{slug}")
    public ResponseEntity<OrgResponse> getBySlug(@PathVariable String slug) {
        return ResponseEntity.ok(organizationService.getBySlug(slug));
    }

    @GetMapping("/orgs")
    public ResponseEntity<List<OrgResponse>> listOrgs(
            @RequestParam(required = false) OrgStatus status,
            @RequestParam(required = false) OrgType type) {
        return ResponseEntity.ok(organizationService.listOrgs(status, type));
    }

    @PatchMapping("/orgs/{id}")
    public ResponseEntity<OrgResponse> updateOrg(
            @PathVariable UUID id,
            @RequestBody UpdateOrgRequest request) {
        return ResponseEntity.ok(organizationService.updateOrg(id, request));
    }

    @PostMapping("/orgs/{id}/suspend")
    public ResponseEntity<Void> suspendOrg(@PathVariable UUID id) {
        organizationService.suspendOrg(id);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/orgs/{id}/activate")
    public ResponseEntity<Void> activateOrg(@PathVariable UUID id) {
        organizationService.activateOrg(id);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/orgs/{id}/subsidiaries")
    public ResponseEntity<List<OrgResponse>> getSubsidiaries(@PathVariable UUID id) {
        return ResponseEntity.ok(organizationService.getSubsidiaries(id));
    }

    // ─── Member endpoints ─────────────────────────────────────────────────────

    @PostMapping("/orgs/{id}/members")
    public ResponseEntity<MemberResponse> inviteMember(
            @PathVariable UUID id,
            @Valid @RequestBody InviteMemberRequest request) {
        MemberResponse response = organizationService.inviteMember(id, request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/orgs/{id}/members")
    public ResponseEntity<List<MemberResponse>> listMembers(
            @PathVariable UUID id,
            @RequestParam(required = false, defaultValue = "false") boolean activeOnly) {
        return ResponseEntity.ok(organizationService.listMembers(id, activeOnly));
    }

    @PatchMapping("/orgs/{id}/members/{memberId}/role")
    public ResponseEntity<Void> updateMemberRole(
            @PathVariable UUID id,
            @PathVariable UUID memberId,
            @Valid @RequestBody UpdateMemberRoleRequest request) {
        organizationService.updateMemberRole(id, memberId, request.role());
        return ResponseEntity.noContent().build();
    }

    @DeleteMapping("/orgs/{id}/members/{memberId}")
    public ResponseEntity<Void> removeMember(
            @PathVariable UUID id,
            @PathVariable UUID memberId) {
        organizationService.removeMember(id, memberId);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/users/{userId}/memberships")
    public ResponseEntity<List<MemberResponse>> getMembershipsByUser(@PathVariable UUID userId) {
        return ResponseEntity.ok(organizationService.getOrgsByUserId(userId));
    }
}
