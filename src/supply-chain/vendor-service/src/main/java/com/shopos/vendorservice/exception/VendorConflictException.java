package com.shopos.vendorservice.exception;

public class VendorConflictException extends RuntimeException {

    public VendorConflictException(String email) {
        super("A vendor with email '" + email + "' already exists");
    }
}
