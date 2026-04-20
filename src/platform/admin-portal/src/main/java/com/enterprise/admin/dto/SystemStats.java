package com.enterprise.admin.dto;

public record SystemStats(
        long totalTenants,
        long activeTenants,
        String uptime,
        String version
) {}
