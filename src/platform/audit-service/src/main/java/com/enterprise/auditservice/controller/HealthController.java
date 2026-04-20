package com.enterprise.auditservice.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

/**
 * Lightweight liveness probe endpoint.
 *
 * <p>Returns HTTP 200 with {@code {"status":"ok"}} whenever the application
 * context is up.  Kubernetes liveness and readiness probes, as well as the
 * platform health-check-service, poll this endpoint.</p>
 */
@RestController
public class HealthController {

    /**
     * Liveness probe.
     *
     * @return 200 OK with body {@code {"status":"ok"}}
     */
    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> healthz() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
