-- V1: Create support tickets and ticket comments tables

CREATE TABLE IF NOT EXISTS support_tickets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number   VARCHAR(20)  NOT NULL UNIQUE,
    customer_id     UUID         NOT NULL,
    order_id        UUID,
    subject         VARCHAR(255) NOT NULL,
    description     TEXT         NOT NULL,
    status          VARCHAR(30)  NOT NULL DEFAULT 'OPEN',
    priority        VARCHAR(20)  NOT NULL DEFAULT 'NORMAL',
    category        VARCHAR(30)  NOT NULL,
    assigned_to     VARCHAR(255),
    resolved_at     TIMESTAMP,
    closed_at       TIMESTAMP,
    created_at      TIMESTAMP    NOT NULL DEFAULT now(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ticket_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id   UUID         NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id   VARCHAR(255) NOT NULL,
    author_type VARCHAR(20)  NOT NULL,
    body        TEXT         NOT NULL,
    internal    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP    NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_tickets_customer_id  ON support_tickets(customer_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status        ON support_tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority      ON support_tickets(priority);
CREATE INDEX IF NOT EXISTS idx_tickets_assigned_to   ON support_tickets(assigned_to);
CREATE INDEX IF NOT EXISTS idx_tickets_created_at    ON support_tickets(created_at);
CREATE INDEX IF NOT EXISTS idx_comments_ticket_id    ON ticket_comments(ticket_id);
CREATE INDEX IF NOT EXISTS idx_comments_internal     ON ticket_comments(ticket_id, internal);
