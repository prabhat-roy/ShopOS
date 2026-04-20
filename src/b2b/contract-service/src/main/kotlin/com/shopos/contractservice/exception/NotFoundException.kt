package com.shopos.contractservice.exception

class NotFoundException(message: String) : RuntimeException(message) {
    constructor(resourceType: String, id: Any) : this("$resourceType not found with id: $id")
}
