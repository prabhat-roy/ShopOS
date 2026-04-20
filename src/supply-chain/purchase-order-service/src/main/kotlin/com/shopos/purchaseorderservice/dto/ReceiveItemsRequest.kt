package com.shopos.purchaseorderservice.dto

import jakarta.validation.Valid
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.NotEmpty
import jakarta.validation.constraints.NotNull
import java.util.UUID

data class ReceiveItemsRequest(

    @field:NotEmpty(message = "At least one item receipt entry is required")
    @field:Valid
    val receipts: List<ItemReceiptEntry>
)

data class ItemReceiptEntry(

    @field:NotNull(message = "itemId is required")
    val itemId: UUID,

    @field:Min(value = 1, message = "receivedQty must be at least 1")
    val receivedQty: Int
)
