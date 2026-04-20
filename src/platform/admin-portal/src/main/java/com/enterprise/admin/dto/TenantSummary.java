package com.enterprise.admin.dto;

public record TenantSummary(
        String id,
        String name,
        String slug,
        String plan,
        String status,
        String ownerEmail
) {}
