package com.shopos.purchaseorderservice.exception

import java.util.UUID

class PurchaseOrderNotFoundException(id: UUID) :
    RuntimeException("Purchase order not found with id: $id")
