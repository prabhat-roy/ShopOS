package com.enterprise.auditservice.repository;

import com.enterprise.auditservice.domain.AuditEvent;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.time.Instant;

/**
 * Spring Data JPA repository for {@link AuditEvent}.
 * All writes are append-only; no delete or update operations are exposed.
 */
@Repository
public interface AuditEventRepository extends JpaRepository<AuditEvent, String> {

    /**
     * Returns a page of audit events triggered by the given actor.
     *
     * @param actorId  the actor identifier (user UUID or service name)
     * @param pageable pagination / sort parameters
     * @return page of matching audit events
     */
    Page<AuditEvent> findByActorId(String actorId, Pageable pageable);

    /**
     * Returns a page of audit events for a specific resource instance.
     *
     * @param resourceType entity type, e.g. "Order"
     * @param resourceId   primary key of the entity
     * @param pageable     pagination / sort parameters
     * @return page of matching audit events
     */
    Page<AuditEvent> findByResourceTypeAndResourceId(String resourceType,
                                                      String resourceId,
                                                      Pageable pageable);

    /**
     * Returns a page of audit events whose {@code occurred_at} falls within
     * the given half-open interval [{@code from}, {@code to}].
     *
     * @param from     lower bound (inclusive)
     * @param to       upper bound (inclusive)
     * @param pageable pagination / sort parameters
     * @return page of matching audit events
     */
    Page<AuditEvent> findByOccurredAtBetween(Instant from, Instant to, Pageable pageable);
}
