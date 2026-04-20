package com.shopos.revenuerecognition

import com.shopos.revenuerecognition.domain.ContractType
import com.shopos.revenuerecognition.domain.RecognitionStatus
import com.shopos.revenuerecognition.domain.RevenueSchedule
import com.shopos.revenuerecognition.dto.CreateScheduleRequest
import com.shopos.revenuerecognition.repository.RevenueScheduleRepository
import com.shopos.revenuerecognition.service.RevenueRecognitionService
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.InjectMocks
import org.mockito.Mock
import org.mockito.junit.jupiter.MockitoExtension
import org.mockito.kotlin.any
import org.mockito.kotlin.whenever
import java.math.BigDecimal
import java.time.Instant
import java.time.LocalDate
import java.util.Optional
import java.util.UUID

@ExtendWith(MockitoExtension::class)
class RevenueRecognitionServiceTest {

    @Mock
    private lateinit var scheduleRepository: RevenueScheduleRepository

    @InjectMocks
    private lateinit var service: RevenueRecognitionService

    private fun sampleSchedule(
        contractType: ContractType = ContractType.SUBSCRIPTION,
        total: BigDecimal = BigDecimal("120.00"),
        start: LocalDate = LocalDate.of(2024, 1, 1),
        end: LocalDate = LocalDate.of(2024, 12, 31),
    ) = RevenueSchedule(
        id = UUID.randomUUID(),
        orderId = "ord-001",
        lineItemId = "li-001",
        contractType = contractType,
        totalAmount = total,
        deferredAmount = total,
        currency = "USD",
        recognitionStartDate = start,
        recognitionEndDate = end,
        status = RecognitionStatus.PENDING,
        createdAt = Instant.now(),
        updatedAt = Instant.now()
    )

    @Test
    fun `createSchedule returns PENDING status`() {
        val schedule = sampleSchedule()
        whenever(scheduleRepository.save(any())).thenReturn(schedule)

        val request = CreateScheduleRequest(
            orderId = "ord-001",
            lineItemId = "li-001",
            contractType = ContractType.SUBSCRIPTION,
            totalAmount = BigDecimal("120.00"),
            currency = "USD",
            recognitionStartDate = LocalDate.of(2024, 1, 1),
            recognitionEndDate = LocalDate.of(2024, 12, 31)
        )

        val response = service.createSchedule(request)
        assertThat(response.status).isEqualTo(RecognitionStatus.PENDING)
        assertThat(response.orderId).isEqualTo("ord-001")
    }

    @Test
    fun `recognizeRevenue for ONE_TIME sets fully recognized`() {
        val schedule = sampleSchedule(contractType = ContractType.ONE_TIME)
        whenever(scheduleRepository.findById(schedule.id!!)).thenReturn(Optional.of(schedule))
        whenever(scheduleRepository.save(any())).thenAnswer { it.arguments[0] as RevenueSchedule }

        val response = service.recognizeRevenue(schedule.id!!, LocalDate.of(2024, 6, 15))
        assertThat(response.status).isEqualTo(RecognitionStatus.FULLY_RECOGNIZED)
        assertThat(response.recognizedAmount).isEqualByComparingTo(BigDecimal("120.00"))
    }

    @Test
    fun `recognizeRevenue subscription at midpoint is IN_PROGRESS`() {
        val schedule = sampleSchedule(
            contractType = ContractType.SUBSCRIPTION,
            start = LocalDate.of(2024, 1, 1),
            end = LocalDate.of(2024, 12, 31)
        )
        whenever(scheduleRepository.findById(schedule.id!!)).thenReturn(Optional.of(schedule))
        whenever(scheduleRepository.save(any())).thenAnswer { it.arguments[0] as RevenueSchedule }

        // Recognizing at exactly halfway through the year
        val response = service.recognizeRevenue(schedule.id!!, LocalDate.of(2024, 7, 2))
        assertThat(response.status).isEqualTo(RecognitionStatus.IN_PROGRESS)
        assertThat(response.recognizedAmount).isGreaterThan(BigDecimal.ZERO)
        assertThat(response.deferredAmount).isGreaterThan(BigDecimal.ZERO)
    }
}
