package com.shopos.accountingservice.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.time.LocalDateTime
import java.util.UUID

@Entity
@Table(name = "accounts")
class Account {

    @Id
    @Column(name = "id", nullable = false, updatable = false)
    var id: UUID = UUID.randomUUID()

    @Column(name = "code", nullable = false, unique = true, length = 50)
    var code: String = ""

    @Column(name = "name", nullable = false, length = 255)
    var name: String = ""

    @Enumerated(EnumType.STRING)
    @Column(name = "type", nullable = false, length = 20)
    var type: AccountType = AccountType.ASSET

    @Column(name = "balance", nullable = false, precision = 19, scale = 4)
    var balance: BigDecimal = BigDecimal.ZERO

    @Column(name = "currency", nullable = false, length = 3)
    var currency: String = "USD"

    @Column(name = "active", nullable = false)
    var active: Boolean = true

    @Column(name = "created_at", nullable = false, updatable = false)
    var createdAt: LocalDateTime = LocalDateTime.now()

    @Column(name = "updated_at", nullable = false)
    var updatedAt: LocalDateTime = LocalDateTime.now()

    @PreUpdate
    fun onUpdate() {
        updatedAt = LocalDateTime.now()
    }
}
