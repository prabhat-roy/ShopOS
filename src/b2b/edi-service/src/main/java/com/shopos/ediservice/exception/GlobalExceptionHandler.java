package com.shopos.ediservice.exception;

import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ProblemDetail;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.FieldError;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import java.net.URI;
import java.time.Instant;
import java.util.List;
import java.util.Map;

/**
 * Centralised exception handler for the EDI service REST API.
 */
@Slf4j
@RestControllerAdvice
public class GlobalExceptionHandler {

    /**
     * Handles bean validation failures (e.g. @NotBlank, @NotNull on request bodies).
     */
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ProblemDetail> handleValidationException(
            MethodArgumentNotValidException ex) {

        List<String> fieldErrors = ex.getBindingResult().getFieldErrors().stream()
                .map(FieldError::getDefaultMessage)
                .toList();

        ProblemDetail problem = ProblemDetail.forStatusAndDetail(
                HttpStatus.BAD_REQUEST,
                "Request validation failed");
        problem.setType(URI.create("https://shopos.io/errors/validation"));
        problem.setTitle("Validation Error");
        problem.setProperty("errors", fieldErrors);
        problem.setProperty("timestamp", Instant.now().toString());

        return ResponseEntity.badRequest().body(problem);
    }

    /**
     * Handles illegal / unsupported EDI format arguments.
     */
    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<ProblemDetail> handleIllegalArgument(IllegalArgumentException ex) {
        log.warn("Illegal argument: {}", ex.getMessage());

        ProblemDetail problem = ProblemDetail.forStatusAndDetail(
                HttpStatus.BAD_REQUEST, ex.getMessage());
        problem.setType(URI.create("https://shopos.io/errors/invalid-input"));
        problem.setTitle("Invalid Input");
        problem.setProperty("timestamp", Instant.now().toString());

        return ResponseEntity.badRequest().body(problem);
    }

    /**
     * Handles requests for features that are planned but not yet released.
     */
    @ExceptionHandler(UnsupportedOperationException.class)
    public ResponseEntity<ProblemDetail> handleUnsupportedOperation(
            UnsupportedOperationException ex) {
        log.warn("Unsupported operation: {}", ex.getMessage());

        ProblemDetail problem = ProblemDetail.forStatusAndDetail(
                HttpStatus.NOT_IMPLEMENTED, ex.getMessage());
        problem.setType(URI.create("https://shopos.io/errors/not-implemented"));
        problem.setTitle("Not Implemented");
        problem.setProperty("timestamp", Instant.now().toString());

        return ResponseEntity.status(HttpStatus.NOT_IMPLEMENTED).body(problem);
    }

    /**
     * Catch-all handler for unexpected runtime errors.
     */
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ProblemDetail> handleGenericException(Exception ex) {
        log.error("Unhandled exception in EDI service", ex);

        ProblemDetail problem = ProblemDetail.forStatusAndDetail(
                HttpStatus.INTERNAL_SERVER_ERROR,
                "An unexpected error occurred. Please contact support.");
        problem.setType(URI.create("https://shopos.io/errors/internal"));
        problem.setTitle("Internal Server Error");
        problem.setProperty("timestamp", Instant.now().toString());

        return ResponseEntity.internalServerError().body(problem);
    }
}
