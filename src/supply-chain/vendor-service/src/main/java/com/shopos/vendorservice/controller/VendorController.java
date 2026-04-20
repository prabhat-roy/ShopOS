package com.shopos.vendorservice.controller;

import com.shopos.vendorservice.domain.VendorStatus;
import com.shopos.vendorservice.dto.CreateVendorRequest;
import com.shopos.vendorservice.dto.UpdateVendorRequest;
import com.shopos.vendorservice.dto.VendorResponse;
import com.shopos.vendorservice.service.VendorService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/vendors")
@RequiredArgsConstructor
public class VendorController {

    private final VendorService vendorService;

    @PostMapping
    public ResponseEntity<VendorResponse> createVendor(@Valid @RequestBody CreateVendorRequest request) {
        VendorResponse response = vendorService.createVendor(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/{id}")
    public ResponseEntity<VendorResponse> getVendor(@PathVariable UUID id) {
        VendorResponse response = vendorService.getVendor(id);
        return ResponseEntity.ok(response);
    }

    @GetMapping
    public ResponseEntity<List<VendorResponse>> listVendors(
            @RequestParam(required = false) VendorStatus status) {
        List<VendorResponse> vendors = vendorService.listVendors(status);
        return ResponseEntity.ok(vendors);
    }

    @PatchMapping("/{id}")
    public ResponseEntity<VendorResponse> updateVendor(
            @PathVariable UUID id,
            @Valid @RequestBody UpdateVendorRequest request) {
        VendorResponse response = vendorService.updateVendor(id, request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/{id}/suspend")
    public ResponseEntity<Void> suspendVendor(@PathVariable UUID id) {
        vendorService.suspendVendor(id);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/{id}/activate")
    public ResponseEntity<Void> activateVendor(@PathVariable UUID id) {
        vendorService.activateVendor(id);
        return ResponseEntity.noContent().build();
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteVendor(@PathVariable UUID id) {
        vendorService.deleteVendor(id);
        return ResponseEntity.noContent().build();
    }
}
