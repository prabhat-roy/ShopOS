package com.shopos.revenuerecognition.exception

import org.slf4j.LoggerFactory
import org.springframework.http.HttpStatus
import org.springframework.http.ProblemDetail
import org.springframework.web.bind.MethodArgumentNotValidException
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice

@RestControllerAdvice
class GlobalExceptionHandler {

    private val log = LoggerFactory.getLogger(javaClass)

    @ExceptionHandler(IllegalArgumentException::class)
    fun handleIllegalArgument(ex: IllegalArgumentException): ProblemDetail {
        log.warn("Resource not found: {}", ex.message)
        val pd = ProblemDetail.forStatusAndDetail(HttpStatus.NOT_FOUND, ex.message ?: "Not found")
        pd.title = "Resource Not Found"
        return pd
    }

    @ExceptionHandler(MethodArgumentNotValidException::class)
    fun handleValidation(ex: MethodArgumentNotValidException): ProblemDetail {
        val detail = ex.bindingResult.fieldErrors
            .joinToString("; ") { "${it.field}: ${it.defaultMessage}" }
            .ifEmpty { "Validation failed" }
        val pd = ProblemDetail.forStatusAndDetail(HttpStatus.BAD_REQUEST, detail)
        pd.title = "Validation Error"
        return pd
    }

    @ExceptionHandler(Exception::class)
    fun handleGeneral(ex: Exception): ProblemDetail {
        log.error("Unhandled exception", ex)
        val pd = ProblemDetail.forStatusAndDetail(
            HttpStatus.INTERNAL_SERVER_ERROR, "An unexpected error occurred"
        )
        pd.title = "Internal Server Error"
        return pd
    }
}
