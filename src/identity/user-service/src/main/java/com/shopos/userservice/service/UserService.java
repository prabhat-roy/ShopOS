package com.shopos.userservice.service;

import com.shopos.userservice.domain.User;
import com.shopos.userservice.domain.UserStatus;
import com.shopos.userservice.dto.CreateUserRequest;
import com.shopos.userservice.dto.UpdateUserRequest;
import com.shopos.userservice.dto.UserResponse;
import com.shopos.userservice.exception.UserNotFoundException;
import com.shopos.userservice.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class UserService {

    private final UserRepository userRepository;

    @Transactional
    public UserResponse createUser(CreateUserRequest request) {
        log.info("Creating user with email={}", request.email());
        User user = User.builder()
                .email(request.email())
                .firstName(request.firstName() != null ? request.firstName() : "")
                .lastName(request.lastName() != null ? request.lastName() : "")
                .phone(request.phone() != null ? request.phone() : "")
                .status(UserStatus.ACTIVE)
                .preferences("{}")
                .build();
        User saved = userRepository.save(user);
        log.info("Created user id={}", saved.getId());
        return UserResponse.from(saved);
    }

    @Transactional(readOnly = true)
    public UserResponse getUser(UUID id) {
        User user = userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException(id));
        return UserResponse.from(user);
    }

    @Transactional(readOnly = true)
    public UserResponse getUserByEmail(String email) {
        User user = userRepository.findByEmail(email)
                .orElseThrow(() -> new UserNotFoundException(email));
        return UserResponse.from(user);
    }

    @Transactional(readOnly = true)
    public Page<UserResponse> listUsers(UserStatus status, Pageable pageable) {
        if (status != null) {
            return userRepository.findByStatus(status, pageable).map(UserResponse::from);
        }
        return userRepository.findAll(pageable).map(UserResponse::from);
    }

    @Transactional
    public UserResponse updateUser(UUID id, UpdateUserRequest request) {
        User user = userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException(id));

        if (request.firstName() != null) {
            user.setFirstName(request.firstName());
        }
        if (request.lastName() != null) {
            user.setLastName(request.lastName());
        }
        if (request.phone() != null) {
            user.setPhone(request.phone());
        }
        if (request.preferences() != null) {
            user.setPreferences(request.preferences());
        }

        User saved = userRepository.save(user);
        log.info("Updated user id={}", saved.getId());
        return UserResponse.from(saved);
    }

    @Transactional
    public void deleteUser(UUID id) {
        User user = userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException(id));
        user.setStatus(UserStatus.DELETED);
        userRepository.save(user);
        log.info("Soft-deleted user id={}", id);
    }

    @Transactional
    public UserResponse suspendUser(UUID id) {
        User user = userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException(id));
        if (user.getStatus() == UserStatus.DELETED) {
            throw new IllegalArgumentException("Cannot suspend a deleted user");
        }
        user.setStatus(UserStatus.SUSPENDED);
        User saved = userRepository.save(user);
        log.info("Suspended user id={}", id);
        return UserResponse.from(saved);
    }

    @Transactional
    public UserResponse activateUser(UUID id) {
        User user = userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException(id));
        if (user.getStatus() == UserStatus.DELETED) {
            throw new IllegalArgumentException("Cannot activate a deleted user");
        }
        user.setStatus(UserStatus.ACTIVE);
        User saved = userRepository.save(user);
        log.info("Activated user id={}", id);
        return UserResponse.from(saved);
    }
}
