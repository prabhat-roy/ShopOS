package com.shopos.revenuerecognition.repository

import com.shopos.revenuerecognition.domain.RecognitionStatus
import com.shopos.revenuerecognition.domain.RevenueSchedule
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository
import java.util.UUID

@Repository
interface RevenueScheduleRepository : JpaRepository<RevenueSchedule, UUID> {

    fun findByOrderId(orderId: String): List<RevenueSchedule>

    fun findByStatus(status: RecognitionStatus): List<RevenueSchedule>
}
