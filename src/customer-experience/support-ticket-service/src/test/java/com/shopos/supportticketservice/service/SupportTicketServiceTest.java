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
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDateTime;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class SupportTicketServiceTest {

    @Mock
    private SupportTicketRepository ticketRepository;

    @Mock
    private TicketCommentRepository commentRepository;

    @InjectMocks
    private SupportTicketService service;

    private UUID customerId;
    private UUID ticketId;
    private SupportTicket sampleTicket;

    @BeforeEach
    void setUp() {
        customerId = UUID.randomUUID();
        ticketId = UUID.randomUUID();
        sampleTicket = SupportTicket.builder()
                .id(ticketId)
                .ticketNumber("TKT-123456")
                .customerId(customerId)
                .subject("Test subject")
                .description("Test description")
                .status(TicketStatus.OPEN)
                .priority(TicketPriority.NORMAL)
                .category(TicketCategory.ORDER)
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();
    }

    // Test 1: createTicket assigns OPEN status and NORMAL priority by default
    @Test
    void createTicket_defaultsToOpenStatusAndNormalPriority() {
        when(ticketRepository.findAll()).thenReturn(Collections.emptyList());
        when(ticketRepository.save(any(SupportTicket.class))).thenAnswer(inv -> inv.getArgument(0));

        CreateTicketRequest request = CreateTicketRequest.builder()
                .customerId(customerId)
                .subject("Order issue")
                .description("My order is delayed")
                .category(TicketCategory.ORDER)
                .build();

        SupportTicket result = service.createTicket(request);

        assertThat(result.getStatus()).isEqualTo(TicketStatus.OPEN);
        assertThat(result.getPriority()).isEqualTo(TicketPriority.NORMAL);
    }

    // Test 2: createTicket uses provided priority when specified
    @Test
    void createTicket_usesPriorityFromRequest() {
        when(ticketRepository.findAll()).thenReturn(Collections.emptyList());
        when(ticketRepository.save(any(SupportTicket.class))).thenAnswer(inv -> inv.getArgument(0));

        CreateTicketRequest request = CreateTicketRequest.builder()
                .customerId(customerId)
                .subject("Urgent payment issue")
                .description("Cannot pay")
                .category(TicketCategory.PAYMENT)
                .priority(TicketPriority.HIGH)
                .build();

        SupportTicket result = service.createTicket(request);

        assertThat(result.getPriority()).isEqualTo(TicketPriority.HIGH);
    }

    // Test 3: createTicket generates a ticket number with TKT- prefix
    @Test
    void createTicket_generatesTicketNumberWithPrefix() {
        when(ticketRepository.findAll()).thenReturn(Collections.emptyList());
        when(ticketRepository.save(any(SupportTicket.class))).thenAnswer(inv -> inv.getArgument(0));

        CreateTicketRequest request = CreateTicketRequest.builder()
                .customerId(customerId)
                .subject("Subject")
                .description("Description")
                .category(TicketCategory.GENERAL)
                .build();

        SupportTicket result = service.createTicket(request);

        assertThat(result.getTicketNumber()).startsWith("TKT-");
        assertThat(result.getTicketNumber()).hasSize(10); // "TKT-" + 6 digits
    }

    // Test 4: getTicket returns ticket when found
    @Test
    void getTicket_returnsTicket_whenFound() {
        when(ticketRepository.findById(ticketId)).thenReturn(Optional.of(sampleTicket));

        SupportTicket result = service.getTicket(ticketId);

        assertThat(result.getId()).isEqualTo(ticketId);
        assertThat(result.getTicketNumber()).isEqualTo("TKT-123456");
    }

    // Test 5: getTicket throws TicketNotFoundException when not found
    @Test
    void getTicket_throwsNotFoundException_whenNotFound() {
        UUID unknownId = UUID.randomUUID();
        when(ticketRepository.findById(unknownId)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> service.getTicket(unknownId))
                .isInstanceOf(TicketNotFoundException.class)
                .hasMessageContaining(unknownId.toString());
    }

    // Test 6: assignTicket sets assignedTo and transitions OPEN -> IN_PROGRESS
    @Test
    void assignTicket_setsAgentAndTransitionsToInProgress() {
        when(ticketRepository.findById(ticketId)).thenReturn(Optional.of(sampleTicket));
        when(ticketRepository.save(any(SupportTicket.class))).thenAnswer(inv -> inv.getArgument(0));

        SupportTicket result = service.assignTicket(ticketId, "agent-007");

        assertThat(result.getAssignedTo()).isEqualTo("agent-007");
        assertThat(result.getStatus()).isEqualTo(TicketStatus.IN_PROGRESS);
    }

    // Test 7: updateStatus to RESOLVED sets resolvedAt timestamp
    @Test
    void updateStatus_toResolved_setsResolvedAt() {
        when(ticketRepository.findById(ticketId)).thenReturn(Optional.of(sampleTicket));
        when(ticketRepository.save(any(SupportTicket.class))).thenAnswer(inv -> inv.getArgument(0));

        SupportTicket result = service.updateStatus(ticketId, TicketStatus.RESOLVED);

        assertThat(result.getStatus()).isEqualTo(TicketStatus.RESOLVED);
        assertThat(result.getResolvedAt()).isNotNull();
    }

    // Test 8: updateStatus to CLOSED sets both resolvedAt and closedAt
    @Test
    void updateStatus_toClosed_setsBothTimestamps() {
        when(ticketRepository.findById(ticketId)).thenReturn(Optional.of(sampleTicket));
        when(ticketRepository.save(any(SupportTicket.class))).thenAnswer(inv -> inv.getArgument(0));

        SupportTicket result = service.updateStatus(ticketId, TicketStatus.CLOSED);

        assertThat(result.getStatus()).isEqualTo(TicketStatus.CLOSED);
        assertThat(result.getResolvedAt()).isNotNull();
        assertThat(result.getClosedAt()).isNotNull();
    }

    // Test 9: escalate sets priority to URGENT
    @Test
    void escalate_setsPriorityToUrgent() {
        when(ticketRepository.findById(ticketId)).thenReturn(Optional.of(sampleTicket));
        when(ticketRepository.save(any(SupportTicket.class))).thenAnswer(inv -> inv.getArgument(0));

        SupportTicket result = service.escalate(ticketId);

        assertThat(result.getPriority()).isEqualTo(TicketPriority.URGENT);
    }

    // Test 10: addComment saves comment with correct fields
    @Test
    void addComment_savesCommentWithCorrectFields() {
        when(ticketRepository.findById(ticketId)).thenReturn(Optional.of(sampleTicket));
        ArgumentCaptor<TicketComment> captor = ArgumentCaptor.forClass(TicketComment.class);
        TicketComment savedComment = TicketComment.builder()
                .id(UUID.randomUUID())
                .ticketId(ticketId)
                .authorId("cust-1")
                .authorType("CUSTOMER")
                .body("Where is my order?")
                .internal(false)
                .build();
        when(commentRepository.save(captor.capture())).thenReturn(savedComment);

        AddCommentRequest request = new AddCommentRequest("cust-1", "CUSTOMER", "Where is my order?", false);
        TicketComment result = service.addComment(ticketId, request);

        TicketComment captured = captor.getValue();
        assertThat(captured.getTicketId()).isEqualTo(ticketId);
        assertThat(captured.getAuthorId()).isEqualTo("cust-1");
        assertThat(captured.getAuthorType()).isEqualTo("CUSTOMER");
        assertThat(captured.getBody()).isEqualTo("Where is my order?");
        assertThat(captured.isInternal()).isFalse();
        assertThat(result.getId()).isNotNull();
    }
}
