package com.shopos.supportticketservice.repository;

import com.shopos.supportticketservice.domain.TicketComment;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface TicketCommentRepository extends JpaRepository<TicketComment, UUID> {

    List<TicketComment> findByTicketIdOrderByCreatedAtAsc(UUID ticketId);

    List<TicketComment> findByTicketIdAndInternalFalseOrderByCreatedAtAsc(UUID ticketId);
}
