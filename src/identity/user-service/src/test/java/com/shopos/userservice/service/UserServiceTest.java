package com.shopos.userservice.service;

import com.shopos.userservice.domain.User;
import com.shopos.userservice.domain.UserStatus;
import com.shopos.userservice.dto.CreateUserRequest;
import com.shopos.userservice.dto.UpdateUserRequest;
import com.shopos.userservice.dto.UserResponse;
import com.shopos.userservice.exception.UserNotFoundException;
import com.shopos.userservice.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.OffsetDateTime;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserServiceTest {

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private UserService userService;

    private UUID userId;
    private User sampleUser;

    @BeforeEach
    void setUp() {
        userId = UUID.randomUUID();
        sampleUser = User.builder()
                .id(userId)
                .email("alice@example.com")
                .firstName("Alice")
                .lastName("Smith")
                .phone("+1-555-0100")
                .status(UserStatus.ACTIVE)
                .preferences("{}")
                .build();
        // Simulate @CreationTimestamp / @UpdateTimestamp
        sampleUser.setCreatedAt(OffsetDateTime.now());
        sampleUser.setUpdatedAt(OffsetDateTime.now());
    }

    // ── createUser ────────────────────────────────────────────────────────────

    @Test
    @DisplayName("createUser: persists entity and returns response DTO")
    void createUser_success() {
        CreateUserRequest req = new CreateUserRequest(
                "alice@example.com", "Alice", "Smith", "+1-555-0100");

        when(userRepository.save(any(User.class))).thenReturn(sampleUser);

        UserResponse response = userService.createUser(req);

        assertThat(response.email()).isEqualTo("alice@example.com");
        assertThat(response.firstName()).isEqualTo("Alice");
        assertThat(response.status()).isEqualTo(UserStatus.ACTIVE);

        ArgumentCaptor<User> captor = ArgumentCaptor.forClass(User.class);
        verify(userRepository).save(captor.capture());
        assertThat(captor.getValue().getStatus()).isEqualTo(UserStatus.ACTIVE);
    }

    // ── getUser ───────────────────────────────────────────────────────────────

    @Test
    @DisplayName("getUser: returns response when user exists")
    void getUser_found() {
        when(userRepository.findById(userId)).thenReturn(Optional.of(sampleUser));

        UserResponse response = userService.getUser(userId);

        assertThat(response.id()).isEqualTo(userId);
        assertThat(response.email()).isEqualTo("alice@example.com");
    }

    @Test
    @DisplayName("getUser: throws UserNotFoundException when user is missing")
    void getUser_notFound() {
        UUID unknown = UUID.randomUUID();
        when(userRepository.findById(unknown)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> userService.getUser(unknown))
                .isInstanceOf(UserNotFoundException.class)
                .hasMessageContaining(unknown.toString());
    }

    // ── updateUser ────────────────────────────────────────────────────────────

    @Test
    @DisplayName("updateUser: applies non-null fields only")
    void updateUser_partialUpdate() {
        UpdateUserRequest req = new UpdateUserRequest("Bob", null, null, null);

        when(userRepository.findById(userId)).thenReturn(Optional.of(sampleUser));
        when(userRepository.save(any(User.class))).thenAnswer(inv -> inv.getArgument(0));

        UserResponse response = userService.updateUser(userId, req);

        assertThat(response.firstName()).isEqualTo("Bob");
        // lastName was null in request, should remain unchanged
        assertThat(response.lastName()).isEqualTo("Smith");
    }

    @Test
    @DisplayName("updateUser: throws UserNotFoundException for unknown id")
    void updateUser_notFound() {
        UUID unknown = UUID.randomUUID();
        when(userRepository.findById(unknown)).thenReturn(Optional.empty());

        UpdateUserRequest req = new UpdateUserRequest("X", null, null, null);

        assertThatThrownBy(() -> userService.updateUser(unknown, req))
                .isInstanceOf(UserNotFoundException.class);
    }

    // ── deleteUser (soft) ─────────────────────────────────────────────────────

    @Test
    @DisplayName("deleteUser: sets status to DELETED without removing the row")
    void deleteUser_softDelete() {
        when(userRepository.findById(userId)).thenReturn(Optional.of(sampleUser));
        when(userRepository.save(any(User.class))).thenAnswer(inv -> inv.getArgument(0));

        userService.deleteUser(userId);

        ArgumentCaptor<User> captor = ArgumentCaptor.forClass(User.class);
        verify(userRepository).save(captor.capture());
        assertThat(captor.getValue().getStatus()).isEqualTo(UserStatus.DELETED);

        // Physical row must NOT be removed
        verify(userRepository, never()).delete(any(User.class));
        verify(userRepository, never()).deleteById(any());
    }

    @Test
    @DisplayName("deleteUser: throws UserNotFoundException for unknown id")
    void deleteUser_notFound() {
        UUID unknown = UUID.randomUUID();
        when(userRepository.findById(unknown)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> userService.deleteUser(unknown))
                .isInstanceOf(UserNotFoundException.class);
    }

    // ── suspendUser ───────────────────────────────────────────────────────────

    @Test
    @DisplayName("suspendUser: sets status to SUSPENDED")
    void suspendUser_success() {
        when(userRepository.findById(userId)).thenReturn(Optional.of(sampleUser));
        when(userRepository.save(any(User.class))).thenAnswer(inv -> inv.getArgument(0));

        UserResponse response = userService.suspendUser(userId);

        assertThat(response.status()).isEqualTo(UserStatus.SUSPENDED);
    }

    @Test
    @DisplayName("suspendUser: throws IllegalArgumentException when user is DELETED")
    void suspendUser_deletedUser() {
        sampleUser.setStatus(UserStatus.DELETED);
        when(userRepository.findById(userId)).thenReturn(Optional.of(sampleUser));

        assertThatThrownBy(() -> userService.suspendUser(userId))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("deleted");
    }

    // ── activateUser ──────────────────────────────────────────────────────────

    @Test
    @DisplayName("activateUser: sets status to ACTIVE")
    void activateUser_success() {
        sampleUser.setStatus(UserStatus.SUSPENDED);
        when(userRepository.findById(userId)).thenReturn(Optional.of(sampleUser));
        when(userRepository.save(any(User.class))).thenAnswer(inv -> inv.getArgument(0));

        UserResponse response = userService.activateUser(userId);

        assertThat(response.status()).isEqualTo(UserStatus.ACTIVE);
    }

    @Test
    @DisplayName("activateUser: throws IllegalArgumentException when user is DELETED")
    void activateUser_deletedUser() {
        sampleUser.setStatus(UserStatus.DELETED);
        when(userRepository.findById(userId)).thenReturn(Optional.of(sampleUser));

        assertThatThrownBy(() -> userService.activateUser(userId))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("deleted");
    }
}
