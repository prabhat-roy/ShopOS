package com.shopos.organizationservice.dto;

import com.shopos.organizationservice.domain.OrgMember;

import java.time.LocalDateTime;
import java.util.UUID;

public record MemberResponse(
        UUID id,
        UUID orgId,
        UUID userId,
        String role,
        String department,
        String jobTitle,
        boolean active,
        LocalDateTime invitedAt,
        LocalDateTime joinedAt,
        LocalDateTime createdAt
) {
    public static MemberResponse from(OrgMember m) {
        return new MemberResponse(
                m.getId(),
                m.getOrgId(),
                m.getUserId(),
                m.getRole(),
                m.getDepartment(),
                m.getJobTitle(),
                m.isActive(),
                m.getInvitedAt(),
                m.getJoinedAt(),
                m.getCreatedAt()
        );
    }
}
