package com.shopos.supportticketservice.controller;

import com.shopos.supportticketservice.domain.TicketPriority;
import com.shopos.supportticketservice.domain.TicketStatus;
import com.shopos.supportticketservice.dto.AddCommentRequest;
import com.shopos.supportticketservice.dto.AssignTicketRequest;
import com.shopos.supportticketservice.dto.CommentResponse;
import com.shopos.supportticketservice.dto.CreateTicketRequest;
import com.shopos.supportticketservice.dto.TicketResponse;
import com.shopos.supportticketservice.dto.UpdateStatusRequest;
import com.shopos.supportticketservice.service.SupportTicketService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/tickets")
@RequiredArgsConstructor
public class SupportTicketController {

    private final SupportTicketService service;

    @PostMapping
    public ResponseEntity<TicketResponse> createTicket(@Valid @RequestBody CreateTicketRequest request) {
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(TicketResponse.from(service.createTicket(request)));
    }

    @GetMapping("/{id}")
    public ResponseEntity<TicketResponse> getTicket(@PathVariable UUID id) {
        return ResponseEntity.ok(TicketResponse.from(service.getTicket(id)));
    }

    @GetMapping
    public ResponseEntity<List<TicketResponse>> listTickets(
            @RequestParam(required = false) UUID customerId,
            @RequestParam(required = false) TicketStatus status,
            @RequestParam(required = false) TicketPriority priority) {
        List<TicketResponse> responses = service.listTickets(customerId, status, priority)
                .stream()
                .map(TicketResponse::from)
                .toList();
        return ResponseEntity.ok(responses);
    }

    @PostMapping("/{id}/assign")
    public ResponseEntity<Void> assignTicket(
            @PathVariable UUID id,
            @Valid @RequestBody AssignTicketRequest request) {
        service.assignTicket(id, request.getAgentId());
        return ResponseEntity.noContent().build();
    }

    @PatchMapping("/{id}/status")
    public ResponseEntity<Void> updateStatus(
            @PathVariable UUID id,
            @Valid @RequestBody UpdateStatusRequest request) {
        service.updateStatus(id, request.getStatus());
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/{id}/escalate")
    public ResponseEntity<Void> escalate(@PathVariable UUID id) {
        service.escalate(id);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/{id}/comments")
    public ResponseEntity<CommentResponse> addComment(
            @PathVariable UUID id,
            @Valid @RequestBody AddCommentRequest request) {
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(CommentResponse.from(service.addComment(id, request)));
    }

    @GetMapping("/{id}/comments")
    public ResponseEntity<List<CommentResponse>> getComments(
            @PathVariable UUID id,
            @RequestParam(defaultValue = "false") boolean includeInternal) {
        List<CommentResponse> responses = service.getComments(id, includeInternal)
                .stream()
                .map(CommentResponse::from)
                .toList();
        return ResponseEntity.ok(responses);
    }

    @GetMapping("/healthz")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
