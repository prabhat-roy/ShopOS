package com.shopos.userservice.dto;

public record UpdateUserRequest(

        String firstName,
        String lastName,
        String phone,

        /** Raw JSON string for user preferences stored as JSONB. */
        String preferences
) {}
