package com.shopos.revenuerecognition.controller

import com.shopos.revenuerecognition.dto.CreateScheduleRequest
import com.shopos.revenuerecognition.dto.ScheduleResponse
import com.shopos.revenuerecognition.service.RevenueRecognitionService
import jakarta.validation.Valid
import org.springframework.format.annotation.DateTimeFormat
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import java.time.LocalDate
import java.util.UUID

@RestController
@RequestMapping("/api/v1/revenue")
class RevenueRecognitionController(
    private val service: RevenueRecognitionService
) {

    @GetMapping("/healthz")
    fun health(): ResponseEntity<Map<String, String>> =
        ResponseEntity.ok(mapOf("status" to "ok"))

    @PostMapping("/schedules")
    fun createSchedule(@Valid @RequestBody request: CreateScheduleRequest): ResponseEntity<ScheduleResponse> =
        ResponseEntity.status(HttpStatus.CREATED).body(service.createSchedule(request))

    @GetMapping("/schedules/{id}")
    fun getSchedule(@PathVariable id: UUID): ResponseEntity<ScheduleResponse> =
        ResponseEntity.ok(service.getSchedule(id))

    @GetMapping("/schedules/order/{orderId}")
    fun getByOrder(@PathVariable orderId: String): ResponseEntity<List<ScheduleResponse>> =
        ResponseEntity.ok(service.getSchedulesByOrder(orderId))

    @PostMapping("/schedules/{id}/recognize")
    fun recognize(
        @PathVariable id: UUID,
        @RequestParam(required = false)
        @DateTimeFormat(iso = DateTimeFormat.ISO.DATE)
        asOfDate: LocalDate?
    ): ResponseEntity<ScheduleResponse> =
        ResponseEntity.ok(service.recognizeRevenue(id, asOfDate ?: LocalDate.now()))
}
