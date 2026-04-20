package com.shopos.vendorservice.exception;

import java.util.UUID;

public class VendorNotFoundException extends RuntimeException {

    public VendorNotFoundException(UUID id) {
        super("Vendor not found with id: " + id);
    }
}
