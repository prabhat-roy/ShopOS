package com.shopos.userservice.domain;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "users")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class User {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "email", nullable = false, unique = true)
    private String email;

    @Column(name = "first_name", nullable = false)
    @Builder.Default
    private String firstName = "";

    @Column(name = "last_name", nullable = false)
    @Builder.Default
    private String lastName = "";

    @Column(name = "phone", nullable = false)
    @Builder.Default
    private String phone = "";

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false)
    @Builder.Default
    private UserStatus status = UserStatus.ACTIVE;

    /**
     * Stored as JSONB in Postgres; mapped as TEXT here.
     * For richer JSONB support, a custom Hibernate type can be added in Phase 2.
     */
    @Column(name = "preferences", nullable = false, columnDefinition = "jsonb")
    @Builder.Default
    private String preferences = "{}";

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private OffsetDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private OffsetDateTime updatedAt;
}
