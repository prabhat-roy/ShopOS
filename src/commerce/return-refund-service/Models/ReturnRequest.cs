namespace ReturnRefundService.Models;

public enum ReturnStatus
{
    Pending,
    Approved,
    Rejected,
    Completed
}

public enum ReturnReason
{
    Defective,
    WrongItem,
    NotAsDescribed,
    ChangedMind,
    Other
}

public class ReturnRequest
{
    public Guid Id { get; set; }
    public string OrderId { get; set; } = "";
    public string CustomerId { get; set; } = "";
    public string ProductId { get; set; } = "";
    public int Quantity { get; set; }
    public ReturnReason Reason { get; set; }
    public string Notes { get; set; } = "";
    public ReturnStatus Status { get; set; } = ReturnStatus.Pending;
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }

    // Navigation property
    public RefundRecord? Refund { get; set; }
}

public class RefundRecord
{
    public Guid Id { get; set; }
    public Guid ReturnRequestId { get; set; }
    public decimal Amount { get; set; }
    public string Currency { get; set; } = "USD";

    /// <summary>
    /// Refund method: "original" (back to payment method) or "store_credit".
    /// </summary>
    public string Method { get; set; } = "original";

    public DateTime ProcessedAt { get; set; }

    // Navigation property
    public ReturnRequest? ReturnRequest { get; set; }
}

// ── Request/Response DTOs ─────────────────────────────────────────────────

public record CreateReturnRequest(
    string OrderId,
    string CustomerId,
    string ProductId,
    int Quantity,
    ReturnReason Reason,
    string Notes = "");

public record UpdateStatusRequest(ReturnStatus Status);

public record ProcessRefundRequest(
    decimal Amount,
    string Method = "original",
    string Currency = "USD");
