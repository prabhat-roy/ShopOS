package com.enterprise.admin.controller;

import com.enterprise.admin.dto.SystemStats;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.lang.management.ManagementFactory;
import java.time.Duration;

@RestController
public class SystemController {

    @Value("${app.version:1.0.0}")
    private String appVersion;

    @GetMapping("/admin/system/stats")
    public SystemStats getStats() {
        long uptimeMillis = ManagementFactory.getRuntimeMXBean().getUptime();
        String uptime = formatUptime(uptimeMillis);

        // Tenant counts are not wired to a live store here; return 0 as a
        // safe default — a future iteration can inject a TenantRepository.
        return new SystemStats(0L, 0L, uptime, appVersion);
    }

    private String formatUptime(long millis) {
        Duration duration = Duration.ofMillis(millis);
        long days    = duration.toDaysPart();
        long hours   = duration.toHoursPart();
        long minutes = duration.toMinutesPart();
        long seconds = duration.toSecondsPart();
        return String.format("%dd %dh %dm %ds", days, hours, minutes, seconds);
    }
}
