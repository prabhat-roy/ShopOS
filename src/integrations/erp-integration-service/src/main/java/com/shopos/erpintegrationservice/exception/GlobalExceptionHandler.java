package com.shopos.erpintegrationservice.exception;

import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ProblemDetail;
import org.springframework.validation.FieldError;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.server.ResponseStatusException;

import java.net.URI;
import java.time.Instant;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Centralised exception handler that translates exceptions into RFC 7807 Problem Detail responses.
 *
 * <p>This handler covers:
 * <ul>
 *   <li>Bean Validation failures → 400 Bad Request with per-field errors.</li>
 *   <li>{@link ResponseStatusException} → passes through the embedded status code.</li>
 *   <li>All other unhandled exceptions → 500 Internal Server Error.</li>
 * </ul>
 */
@Slf4j
@RestControllerAdvice
public class GlobalExceptionHandler {

    /**
     * Handles @Valid / @Validated constraint violations from request bodies.
     *
     * @param ex The exception raised by Spring's validation framework.
     * @return 400 Bad Request with a list of field-level error messages.
     */
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ProblemDetail handleValidationErrors(MethodArgumentNotValidException ex) {
        List<String> fieldErrors = ex.getBindingResult().getFieldErrors().stream()
                .map(FieldError::getDefaultMessage)
                .toList();

        ProblemDetail problem = ProblemDetail.forStatus(HttpStatus.BAD_REQUEST);
        problem.setType(URI.create("urn:shopos:erp:validation-error"));
        problem.setTitle("Validation Failed");
        problem.setDetail("One or more request fields are invalid.");
        problem.setProperty("errors", fieldErrors);
        problem.setProperty("timestamp", Instant.now().toString());
        return problem;
    }

    /**
     * Passes through {@link ResponseStatusException}s raised by controllers and services.
     *
     * @param ex The exception containing the desired HTTP status and reason.
     * @return A Problem Detail response matching the exception's status.
     */
    @ExceptionHandler(ResponseStatusException.class)
    public ProblemDetail handleResponseStatusException(ResponseStatusException ex) {
        ProblemDetail problem = ProblemDetail.forStatus(ex.getStatusCode());
        problem.setType(URI.create("urn:shopos:erp:http-error"));
        problem.setTitle(ex.getReason() != null ? ex.getReason() : "HTTP Error");
        problem.setDetail(ex.getMessage());
        problem.setProperty("timestamp", Instant.now().toString());
        return problem;
    }

    /**
     * Catch-all handler for unexpected runtime exceptions.
     *
     * @param ex The unexpected exception.
     * @return 500 Internal Server Error without exposing internal stack details.
     */
    @ExceptionHandler(Exception.class)
    public ProblemDetail handleUnexpectedException(Exception ex) {
        log.error("Unhandled exception in erp-integration-service: {}", ex.getMessage(), ex);
        ProblemDetail problem = ProblemDetail.forStatus(HttpStatus.INTERNAL_SERVER_ERROR);
        problem.setType(URI.create("urn:shopos:erp:internal-error"));
        problem.setTitle("Internal Server Error");
        problem.setDetail("An unexpected error occurred. Please contact support.");
        problem.setProperty("timestamp", Instant.now().toString());
        return problem;
    }
}
