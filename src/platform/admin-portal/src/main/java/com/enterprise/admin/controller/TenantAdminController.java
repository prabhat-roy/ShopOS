package com.enterprise.admin.controller;

import com.enterprise.admin.dto.TenantSummary;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.web.reactive.function.client.WebClientException;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/admin/tenants")
public class TenantAdminController {

    private final WebClient webClient;

    @Value("${services.tenant-url}")
    private String tenantServiceUrl;

    public TenantAdminController(WebClient.Builder webClientBuilder) {
        this.webClient = webClientBuilder.build();
    }

    @GetMapping
    public ResponseEntity<?> listTenants() {
        try {
            List<TenantSummary> tenants = webClient.get()
                .uri(tenantServiceUrl + "/tenants")
                .retrieve()
                .bodyToMono(new ParameterizedTypeReference<List<TenantSummary>>() {})
                .block();
            return ResponseEntity.ok(tenants);
        } catch (WebClientException | IllegalStateException ex) {
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                .body(Map.of("error", "Tenant service unavailable", "detail", ex.getMessage()));
        }
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> getTenant(@PathVariable String id) {
        try {
            TenantSummary tenant = webClient.get()
                .uri(tenantServiceUrl + "/tenants/" + id)
                .retrieve()
                .bodyToMono(TenantSummary.class)
                .block();
            return ResponseEntity.ok(tenant);
        } catch (WebClientException | IllegalStateException ex) {
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                .body(Map.of("error", "Tenant service unavailable", "detail", ex.getMessage()));
        }
    }

    @PostMapping("/{id}/suspend")
    public ResponseEntity<?> suspendTenant(@PathVariable String id) {
        return patchTenantStatus(id, "suspended");
    }

    @PostMapping("/{id}/activate")
    public ResponseEntity<?> activateTenant(@PathVariable String id) {
        return patchTenantStatus(id, "active");
    }

    private ResponseEntity<?> patchTenantStatus(String id, String status) {
        try {
            TenantSummary updated = webClient.patch()
                .uri(tenantServiceUrl + "/tenants/" + id)
                .bodyValue(Map.of("status", status))
                .retrieve()
                .bodyToMono(TenantSummary.class)
                .block();
            return ResponseEntity.ok(updated);
        } catch (WebClientException | IllegalStateException ex) {
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                .body(Map.of("error", "Tenant service unavailable", "detail", ex.getMessage()));
        }
    }
}
