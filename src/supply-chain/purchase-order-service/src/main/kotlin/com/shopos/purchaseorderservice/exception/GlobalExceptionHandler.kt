package com.shopos.purchaseorderservice.exception

import org.springframework.http.HttpStatus
import org.springframework.http.ProblemDetail
import org.springframework.validation.FieldError
import org.springframework.web.bind.MethodArgumentNotValidException
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice
import java.net.URI
import java.time.Instant

@RestControllerAdvice
class GlobalExceptionHandler {

    @ExceptionHandler(PurchaseOrderNotFoundException::class)
    fun handleNotFound(ex: PurchaseOrderNotFoundException): ProblemDetail {
        val problem = ProblemDetail.forStatusAndDetail(HttpStatus.NOT_FOUND, ex.message ?: "Not found")
        problem.title = "Purchase Order Not Found"
        problem.type = URI.create("https://shopos.com/errors/purchase-order-not-found")
        problem.setProperty("timestamp", Instant.now())
        return problem
    }

    @ExceptionHandler(InvalidPOTransitionException::class)
    fun handleInvalidTransition(ex: InvalidPOTransitionException): ProblemDetail {
        val problem = ProblemDetail.forStatusAndDetail(HttpStatus.UNPROCESSABLE_ENTITY, ex.message ?: "Invalid transition")
        problem.title = "Invalid Status Transition"
        problem.type = URI.create("https://shopos.com/errors/invalid-po-transition")
        problem.setProperty("timestamp", Instant.now())
        return problem
    }

    @ExceptionHandler(IllegalArgumentException::class)
    fun handleIllegalArgument(ex: IllegalArgumentException): ProblemDetail {
        val problem = ProblemDetail.forStatusAndDetail(HttpStatus.BAD_REQUEST, ex.message ?: "Bad request")
        problem.title = "Bad Request"
        problem.type = URI.create("https://shopos.com/errors/bad-request")
        problem.setProperty("timestamp", Instant.now())
        return problem
    }

    @ExceptionHandler(MethodArgumentNotValidException::class)
    fun handleValidation(ex: MethodArgumentNotValidException): ProblemDetail {
        val fieldErrors = ex.bindingResult.fieldErrors
            .associate { fe: FieldError -> fe.field to (fe.defaultMessage ?: "Invalid value") }
        val problem = ProblemDetail.forStatusAndDetail(HttpStatus.BAD_REQUEST, "Request validation failed")
        problem.title = "Validation Error"
        problem.type = URI.create("https://shopos.com/errors/validation-error")
        problem.setProperty("timestamp", Instant.now())
        problem.setProperty("fieldErrors", fieldErrors)
        return problem
    }

    @ExceptionHandler(Exception::class)
    fun handleGeneric(ex: Exception): ProblemDetail {
        val problem = ProblemDetail.forStatusAndDetail(HttpStatus.INTERNAL_SERVER_ERROR, "An unexpected error occurred")
        problem.title = "Internal Server Error"
        problem.type = URI.create("https://shopos.com/errors/internal-server-error")
        problem.setProperty("timestamp", Instant.now())
        return problem
    }
}
