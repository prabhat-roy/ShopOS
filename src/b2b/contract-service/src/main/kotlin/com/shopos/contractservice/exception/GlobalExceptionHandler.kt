package com.shopos.contractservice.exception

import org.springframework.http.HttpStatus
import org.springframework.http.ProblemDetail
import org.springframework.validation.FieldError
import org.springframework.web.bind.MethodArgumentNotValidException
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice
import java.net.URI

@RestControllerAdvice
class GlobalExceptionHandler {

    @ExceptionHandler(NotFoundException::class)
    fun handleNotFoundException(ex: NotFoundException): ProblemDetail {
        val problem = ProblemDetail.forStatusAndDetail(HttpStatus.NOT_FOUND, ex.message ?: "Not found")
        problem.title = "Resource Not Found"
        problem.type = URI.create("https://shopos.com/errors/not-found")
        return problem
    }

    @ExceptionHandler(IllegalStateException::class)
    fun handleIllegalStateException(ex: IllegalStateException): ProblemDetail {
        val problem = ProblemDetail.forStatusAndDetail(HttpStatus.UNPROCESSABLE_ENTITY, ex.message ?: "Invalid state")
        problem.title = "Invalid State Transition"
        problem.type = URI.create("https://shopos.com/errors/invalid-state")
        return problem
    }

    @ExceptionHandler(MethodArgumentNotValidException::class)
    fun handleValidationException(ex: MethodArgumentNotValidException): ProblemDetail {
        val errors = ex.bindingResult.fieldErrors.associate { fe: FieldError ->
            fe.field to (fe.defaultMessage ?: "invalid")
        }
        val problem = ProblemDetail.forStatusAndDetail(
            HttpStatus.BAD_REQUEST, "Validation failed for one or more fields"
        )
        problem.title = "Validation Error"
        problem.type = URI.create("https://shopos.com/errors/validation")
        problem.setProperty("fieldErrors", errors)
        return problem
    }

    @ExceptionHandler(Exception::class)
    fun handleGenericException(ex: Exception): ProblemDetail {
        val problem = ProblemDetail.forStatusAndDetail(
            HttpStatus.INTERNAL_SERVER_ERROR, "An unexpected error occurred"
        )
        problem.title = "Internal Server Error"
        problem.type = URI.create("https://shopos.com/errors/internal")
        return problem
    }
}
