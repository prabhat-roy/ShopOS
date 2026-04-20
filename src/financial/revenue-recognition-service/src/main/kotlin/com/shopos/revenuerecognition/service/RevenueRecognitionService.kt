package com.shopos.revenuerecognition.service

import com.shopos.revenuerecognition.domain.ContractType
import com.shopos.revenuerecognition.domain.RecognitionStatus
import com.shopos.revenuerecognition.domain.RevenueSchedule
import com.shopos.revenuerecognition.dto.CreateScheduleRequest
import com.shopos.revenuerecognition.dto.ScheduleResponse
import com.shopos.revenuerecognition.repository.RevenueScheduleRepository
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.math.RoundingMode
import java.time.LocalDate
import java.time.temporal.ChronoUnit
import java.util.UUID

@Service
class RevenueRecognitionService(
    private val scheduleRepository: RevenueScheduleRepository
) {

    private val log = LoggerFactory.getLogger(javaClass)

    @Transactional
    fun createSchedule(request: CreateScheduleRequest): ScheduleResponse {
        val schedule = RevenueSchedule(
            orderId = request.orderId,
            lineItemId = request.lineItemId,
            contractType = request.contractType,
            totalAmount = request.totalAmount,
            deferredAmount = request.totalAmount,
            currency = request.currency,
            recognitionStartDate = request.recognitionStartDate,
            recognitionEndDate = request.recognitionEndDate,
            status = RecognitionStatus.PENDING
        )
        val saved = scheduleRepository.save(schedule)
        log.info("Revenue schedule created id={} orderId={} type={}", saved.id, saved.orderId, saved.contractType)
        return saved.toResponse()
    }

    @Transactional(readOnly = true)
    fun getSchedule(id: UUID): ScheduleResponse =
        scheduleRepository.findById(id)
            .orElseThrow { IllegalArgumentException("Schedule not found: $id") }
            .toResponse()

    @Transactional(readOnly = true)
    fun getSchedulesByOrder(orderId: String): List<ScheduleResponse> =
        scheduleRepository.findByOrderId(orderId).map { it.toResponse() }

    /**
     * Runs the periodic recognition pass for a given schedule.
     * Applies straight-line (subscription), point-in-time (one-time), or
     * redemption-based (gift card) recognition as appropriate.
     */
    @Transactional
    fun recognizeRevenue(id: UUID, asOfDate: LocalDate): ScheduleResponse {
        val schedule = scheduleRepository.findById(id)
            .orElseThrow { IllegalArgumentException("Schedule not found: $id") }

        val recognized = when (schedule.contractType) {
            ContractType.ONE_TIME -> schedule.totalAmount
            ContractType.SUBSCRIPTION, ContractType.MULTI_ELEMENT ->
                computeStraightLine(schedule, asOfDate)
            ContractType.GIFT_CARD ->
                // Gift cards recognized at redemption; caller marks fully recognized
                schedule.totalAmount
        }

        schedule.recognizedAmount = recognized.min(schedule.totalAmount)
        schedule.deferredAmount = (schedule.totalAmount - schedule.recognizedAmount)
            .max(BigDecimal.ZERO)
        schedule.status = if (schedule.deferredAmount.compareTo(BigDecimal.ZERO) == 0) {
            RecognitionStatus.FULLY_RECOGNIZED
        } else {
            RecognitionStatus.IN_PROGRESS
        }

        val saved = scheduleRepository.save(schedule)
        log.info(
            "Revenue recognized id={} recognized={} deferred={} status={}",
            saved.id, saved.recognizedAmount, saved.deferredAmount, saved.status
        )
        return saved.toResponse()
    }

    private fun computeStraightLine(schedule: RevenueSchedule, asOfDate: LocalDate): BigDecimal {
        val totalDays = ChronoUnit.DAYS.between(
            schedule.recognitionStartDate, schedule.recognitionEndDate
        ).coerceAtLeast(1)

        val elapsedDays = ChronoUnit.DAYS.between(
            schedule.recognitionStartDate,
            minOf(asOfDate, schedule.recognitionEndDate)
        ).coerceAtLeast(0)

        return schedule.totalAmount
            .multiply(BigDecimal(elapsedDays))
            .divide(BigDecimal(totalDays), 4, RoundingMode.HALF_UP)
    }

    private fun RevenueSchedule.toResponse() = ScheduleResponse(
        id = id,
        orderId = orderId,
        lineItemId = lineItemId,
        contractType = contractType,
        totalAmount = totalAmount,
        recognizedAmount = recognizedAmount,
        deferredAmount = deferredAmount,
        currency = currency,
        recognitionStartDate = recognitionStartDate,
        recognitionEndDate = recognitionEndDate,
        status = status,
        createdAt = createdAt,
        updatedAt = updatedAt
    )
}
