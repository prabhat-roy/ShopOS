package com.shopos.purchaseorderservice.exception

import com.shopos.purchaseorderservice.domain.POStatus

class InvalidPOTransitionException(from: POStatus, to: POStatus) :
    RuntimeException("Cannot transition purchase order from $from to $to")
