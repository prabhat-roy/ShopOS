package com.shopos.ediservice.controller;

import com.shopos.ediservice.dto.EdiResponse;
import com.shopos.ediservice.dto.GenerateRequest;
import com.shopos.ediservice.dto.ParseRequest;
import com.shopos.ediservice.service.EdiService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

/**
 * REST controller exposing EDI parse, generate and validate operations.
 */
@Slf4j
@RestController
@RequestMapping("/edi")
@RequiredArgsConstructor
public class EdiController {

    private final EdiService ediService;

    /**
     * Parses a raw EDI message and returns the structured business document.
     *
     * <p>POST /edi/parse
     */
    @PostMapping(value = "/parse",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<EdiResponse> parse(@Valid @RequestBody ParseRequest request) {
        log.info("POST /edi/parse — format={}", request.format());
        EdiResponse response = ediService.parse(request);
        return response.success()
                ? ResponseEntity.ok(response)
                : ResponseEntity.unprocessableEntity().body(response);
    }

    /**
     * Generates a raw EDI string from a business document JSON.
     *
     * <p>POST /edi/generate
     */
    @PostMapping(value = "/generate",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.TEXT_PLAIN_VALUE)
    public ResponseEntity<String> generate(@Valid @RequestBody GenerateRequest request) {
        log.info("POST /edi/generate — format={}, documentType={}", request.format(), request.documentType());
        String ediContent = ediService.generate(request);
        return ResponseEntity.ok()
                .contentType(MediaType.TEXT_PLAIN)
                .body(ediContent);
    }

    /**
     * Validates an EDI message for structural correctness without fully parsing it.
     *
     * <p>POST /edi/validate
     */
    @PostMapping(value = "/validate",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, Object>> validate(@Valid @RequestBody ParseRequest request) {
        log.info("POST /edi/validate — format={}", request.format());
        List<String> errors = ediService.validateFormat(request.content(), request.format());

        Map<String, Object> result = Map.of(
                "valid", errors.isEmpty(),
                "format", request.format().name(),
                "errors", errors
        );

        return errors.isEmpty()
                ? ResponseEntity.ok(result)
                : ResponseEntity.unprocessableEntity().body(result);
    }

    /**
     * Health check endpoint.
     *
     * <p>GET /healthz
     */
    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> healthz() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
