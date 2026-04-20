package com.shopos.accountingservice.repository

import com.shopos.accountingservice.domain.Account
import com.shopos.accountingservice.domain.AccountType
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository
import java.util.Optional
import java.util.UUID

@Repository
interface AccountRepository : JpaRepository<Account, UUID> {

    fun findByCode(code: String): Optional<Account>

    fun findByType(type: AccountType): List<Account>

    fun findByActiveTrue(): List<Account>

    fun findByTypeAndActiveTrue(type: AccountType): List<Account>

    fun existsByCode(code: String): Boolean
}
