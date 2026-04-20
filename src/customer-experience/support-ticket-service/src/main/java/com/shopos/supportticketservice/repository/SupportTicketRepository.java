package com.shopos.supportticketservice.repository;

import com.shopos.supportticketservice.domain.SupportTicket;
import com.shopos.supportticketservice.domain.TicketPriority;
import com.shopos.supportticketservice.domain.TicketStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Repository
public interface SupportTicketRepository extends JpaRepository<SupportTicket, UUID> {

    List<SupportTicket> findByCustomerId(UUID customerId);

    List<SupportTicket> findByStatus(TicketStatus status);

    List<SupportTicket> findByPriority(TicketPriority priority);

    List<SupportTicket> findByAssignedTo(String assignedTo);

    List<SupportTicket> findByCreatedAtBetween(LocalDateTime from, LocalDateTime to);

    List<SupportTicket> findByCustomerIdAndStatus(UUID customerId, TicketStatus status);

    List<SupportTicket> findByCustomerIdAndPriority(UUID customerId, TicketPriority priority);

    List<SupportTicket> findByCustomerIdAndStatusAndPriority(UUID customerId, TicketStatus status, TicketPriority priority);
}
