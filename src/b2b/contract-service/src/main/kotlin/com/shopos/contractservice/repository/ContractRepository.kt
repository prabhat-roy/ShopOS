package com.shopos.contractservice.repository

import com.shopos.contractservice.domain.Contract
import com.shopos.contractservice.domain.ContractStatus
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository
import java.time.LocalDate
import java.util.UUID

@Repository
interface ContractRepository : JpaRepository<Contract, UUID> {

    fun findByOrgId(orgId: UUID): List<Contract>

    fun findByStatus(status: ContractStatus): List<Contract>

    fun findByOrgIdAndStatus(orgId: UUID, status: ContractStatus): List<Contract>

    fun findByEndDateBeforeAndStatus(date: LocalDate, status: ContractStatus): List<Contract>
}
