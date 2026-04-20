package com.shopos.userservice.dto;

import com.shopos.userservice.domain.User;
import com.shopos.userservice.domain.UserStatus;

import java.time.OffsetDateTime;
import java.util.UUID;

public record UserResponse(

        UUID id,
        String email,
        String firstName,
        String lastName,
        String phone,
        UserStatus status,
        OffsetDateTime createdAt
) {
    public static UserResponse from(User user) {
        return new UserResponse(
                user.getId(),
                user.getEmail(),
                user.getFirstName(),
                user.getLastName(),
                user.getPhone(),
                user.getStatus(),
                user.getCreatedAt()
        );
    }
}
