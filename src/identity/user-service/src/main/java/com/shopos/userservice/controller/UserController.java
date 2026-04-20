package com.shopos.userservice.controller;

import com.shopos.userservice.domain.UserStatus;
import com.shopos.userservice.dto.CreateUserRequest;
import com.shopos.userservice.dto.UpdateUserRequest;
import com.shopos.userservice.dto.UserResponse;
import com.shopos.userservice.service.UserService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@RestController
@RequestMapping("/users")
@RequiredArgsConstructor
public class UserController {

    private final UserService userService;

    /**
     * POST /users — create a new user account.
     * Returns 201 Created with the created user body.
     */
    @PostMapping
    public ResponseEntity<UserResponse> createUser(@Valid @RequestBody CreateUserRequest request) {
        UserResponse created = userService.createUser(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    /**
     * GET /users/{id} — fetch a single user by UUID.
     */
    @GetMapping("/{id}")
    public ResponseEntity<UserResponse> getUser(@PathVariable UUID id) {
        return ResponseEntity.ok(userService.getUser(id));
    }

    /**
     * GET /users?status=ACTIVE&page=0&size=20 — paginated list, optional status filter.
     */
    @GetMapping
    public ResponseEntity<Page<UserResponse>> listUsers(
            @RequestParam(required = false) UserStatus status,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size
    ) {
        Pageable pageable = PageRequest.of(page, size, Sort.by("createdAt").descending());
        return ResponseEntity.ok(userService.listUsers(status, pageable));
    }

    /**
     * PATCH /users/{id} — partial update of profile fields.
     */
    @PatchMapping("/{id}")
    public ResponseEntity<UserResponse> updateUser(
            @PathVariable UUID id,
            @RequestBody UpdateUserRequest request
    ) {
        return ResponseEntity.ok(userService.updateUser(id, request));
    }

    /**
     * DELETE /users/{id} — soft-delete (sets status to DELETED).
     * Returns 204 No Content.
     */
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable UUID id) {
        userService.deleteUser(id);
        return ResponseEntity.noContent().build();
    }

    /**
     * POST /users/{id}/suspend — set user status to SUSPENDED.
     */
    @PostMapping("/{id}/suspend")
    public ResponseEntity<UserResponse> suspendUser(@PathVariable UUID id) {
        return ResponseEntity.ok(userService.suspendUser(id));
    }

    /**
     * POST /users/{id}/activate — set user status to ACTIVE.
     */
    @PostMapping("/{id}/activate")
    public ResponseEntity<UserResponse> activateUser(@PathVariable UUID id) {
        return ResponseEntity.ok(userService.activateUser(id));
    }
}
