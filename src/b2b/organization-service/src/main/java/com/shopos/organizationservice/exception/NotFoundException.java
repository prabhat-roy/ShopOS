package com.shopos.organizationservice.exception;

public class NotFoundException extends RuntimeException {

    public NotFoundException(String message) {
        super(message);
    }

    public NotFoundException(String resourceType, Object id) {
        super(resourceType + " not found with id: " + id);
    }
}
