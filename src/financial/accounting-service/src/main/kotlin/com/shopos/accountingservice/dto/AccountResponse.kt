package com.shopos.accountingservice.dto

import com.shopos.accountingservice.domain.Account
import com.shopos.accountingservice.domain.AccountType
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

data class AccountResponse(
    val id: UUID,
    val code: String,
    val name: String,
    val type: AccountType,
    val balance: BigDecimal,
    val currency: String,
    val active: Boolean,
    val createdAt: LocalDateTime,
    val updatedAt: LocalDateTime
) {
    companion object {
        fun from(account: Account): AccountResponse = AccountResponse(
            id = account.id,
            code = account.code,
            name = account.name,
            type = account.type,
            balance = account.balance,
            currency = account.currency,
            active = account.active,
            createdAt = account.createdAt,
            updatedAt = account.updatedAt
        )
    }
}
