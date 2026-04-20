package com.shopos.supportticketservice.service;

import com.shopos.supportticketservice.domain.SupportTicket;
import com.shopos.supportticketservice.domain.TicketCategory;
import com.shopos.supportticketservice.domain.TicketComment;
import com.shopos.supportticketservice.domain.TicketPriority;
import com.shopos.supportticketservice.domain.TicketStatus;
import com.shopos.supportticketservice.dto.AddCommentRequest;
import com.shopos.supportticketservice.dto.CreateTicketRequest;
import com.shopos.supportticketservice.exception.TicketNotFoundException;
import com.shopos.supportticketservice.repository.SupportTicketRepository;
import com.shopos.supportticketservice.repository.TicketCommentRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Random;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class SupportTicketService {

    private final SupportTicketRepository ticketRepository;
    private final TicketCommentRepository commentRepository;
    private final Random random = new Random();

    @Transactional
    public SupportTicket createTicket(CreateTicketRequest request) {
        SupportTicket ticket = SupportTicket.builder()
                .ticketNumber(generateTicketNumber())
                .customerId(request.getCustomerId())
                .orderId(request.getOrderId())
                .subject(request.getSubject())
                .description(request.getDescription())
                .status(TicketStatus.OPEN)
                .priority(request.getPriority() != null ? request.getPriority() : TicketPriority.NORMAL)
                .category(request.getCategory())
                .build();
        return ticketRepository.save(ticket);
    }

    @Transactional(readOnly = true)
    public SupportTicket getTicket(UUID id) {
        return ticketRepository.findById(id)
                .orElseThrow(() -> new TicketNotFoundException(id));
    }

    @Transactional(readOnly = true)
    public List<SupportTicket> listTickets(UUID customerId, TicketStatus status, TicketPriority priority) {
        if (customerId != null && status != null && priority != null) {
            return ticketRepository.findByCustomerIdAndStatusAndPriority(customerId, status, priority);
        } else if (customerId != null && status != null) {
            return ticketRepository.findByCustomerIdAndStatus(customerId, status);
        } else if (customerId != null && priority != null) {
            return ticketRepository.findByCustomerIdAndPriority(customerId, priority);
        } else if (customerId != null) {
            return ticketRepository.findByCustomerId(customerId);
        } else if (status != null) {
            return ticketRepository.findByStatus(status);
        } else if (priority != null) {
            return ticketRepository.findByPriority(priority);
        }
        return ticketRepository.findAll();
    }

    @Transactional
    public SupportTicket assignTicket(UUID id, String agentId) {
        SupportTicket ticket = getTicket(id);
        ticket.setAssignedTo(agentId);
        if (ticket.getStatus() == TicketStatus.OPEN) {
            ticket.setStatus(TicketStatus.IN_PROGRESS);
        }
        return ticketRepository.save(ticket);
    }

    @Transactional
    public SupportTicket updateStatus(UUID id, TicketStatus newStatus) {
        SupportTicket ticket = getTicket(id);
        ticket.setStatus(newStatus);
        LocalDateTime now = LocalDateTime.now();
        if (newStatus == TicketStatus.RESOLVED && ticket.getResolvedAt() == null) {
            ticket.setResolvedAt(now);
        }
        if (newStatus == TicketStatus.CLOSED) {
            if (ticket.getResolvedAt() == null) {
                ticket.setResolvedAt(now);
            }
            ticket.setClosedAt(now);
        }
        return ticketRepository.save(ticket);
    }

    @Transactional
    public TicketComment addComment(UUID ticketId, AddCommentRequest request) {
        // Verify ticket exists
        getTicket(ticketId);
        TicketComment comment = TicketComment.builder()
                .ticketId(ticketId)
                .authorId(request.getAuthorId())
                .authorType(request.getAuthorType())
                .body(request.getBody())
                .internal(request.isInternal())
                .build();
        return commentRepository.save(comment);
    }

    @Transactional(readOnly = true)
    public List<TicketComment> getComments(UUID ticketId, boolean includeInternal) {
        // Verify ticket exists
        getTicket(ticketId);
        if (includeInternal) {
            return commentRepository.findByTicketIdOrderByCreatedAtAsc(ticketId);
        }
        return commentRepository.findByTicketIdAndInternalFalseOrderByCreatedAtAsc(ticketId);
    }

    @Transactional
    public SupportTicket escalate(UUID id) {
        SupportTicket ticket = getTicket(id);
        ticket.setPriority(TicketPriority.URGENT);
        if (ticket.getStatus() == TicketStatus.OPEN) {
            ticket.setStatus(TicketStatus.IN_PROGRESS);
        }
        return ticketRepository.save(ticket);
    }

    private String generateTicketNumber() {
        String number;
        do {
            int n = random.nextInt(900000) + 100000;
            number = "TKT-" + n;
            final String candidate = number;
            if (ticketRepository.findAll().stream().noneMatch(t -> candidate.equals(t.getTicketNumber()))) {
                return number;
            }
        } while (true);
    }
}
