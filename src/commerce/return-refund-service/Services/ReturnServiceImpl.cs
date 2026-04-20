using Microsoft.EntityFrameworkCore;
using ReturnRefundService.Data;
using ReturnRefundService.Models;

namespace ReturnRefundService.Services;

public class ReturnServiceImpl : IReturnService
{
    private readonly AppDbContext _db;
    private readonly ILogger<ReturnServiceImpl> _logger;

    public ReturnServiceImpl(AppDbContext db, ILogger<ReturnServiceImpl> logger)
    {
        _db = db;
        _logger = logger;
    }

    public async Task<ReturnRequest?> GetReturnAsync(Guid id)
    {
        _logger.LogDebug("Fetching return request {Id}", id);

        return await _db.ReturnRequests
            .Include(r => r.Refund)
            .FirstOrDefaultAsync(r => r.Id == id);
    }

    public async Task<IReadOnlyList<ReturnRequest>> ListReturnsAsync(string customerId)
    {
        if (string.IsNullOrWhiteSpace(customerId))
            throw new ArgumentException("CustomerId is required.", nameof(customerId));

        _logger.LogDebug("Listing returns for customer {CustomerId}", customerId);

        return await _db.ReturnRequests
            .Include(r => r.Refund)
            .Where(r => r.CustomerId == customerId)
            .OrderByDescending(r => r.CreatedAt)
            .ToListAsync();
    }

    public async Task<ReturnRequest> CreateReturnAsync(CreateReturnRequest request)
    {
        if (string.IsNullOrWhiteSpace(request.OrderId))
            throw new ArgumentException("OrderId is required.", nameof(request));
        if (string.IsNullOrWhiteSpace(request.CustomerId))
            throw new ArgumentException("CustomerId is required.", nameof(request));
        if (string.IsNullOrWhiteSpace(request.ProductId))
            throw new ArgumentException("ProductId is required.", nameof(request));
        if (request.Quantity <= 0)
            throw new ArgumentException("Quantity must be greater than zero.", nameof(request));

        _logger.LogInformation(
            "Creating return for order {OrderId}, product {ProductId}, customer {CustomerId}",
            request.OrderId, request.ProductId, request.CustomerId);

        var returnRequest = new ReturnRequest
        {
            Id = Guid.NewGuid(),
            OrderId = request.OrderId,
            CustomerId = request.CustomerId,
            ProductId = request.ProductId,
            Quantity = request.Quantity,
            Reason = request.Reason,
            Notes = request.Notes ?? "",
            Status = ReturnStatus.Pending,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };

        _db.ReturnRequests.Add(returnRequest);
        await _db.SaveChangesAsync();

        return returnRequest;
    }

    public async Task<ReturnRequest> UpdateStatusAsync(Guid id, ReturnStatus status)
    {
        _logger.LogInformation("Updating status of return {Id} to {Status}", id, status);

        var returnRequest = await _db.ReturnRequests.FindAsync(id)
            ?? throw new KeyNotFoundException($"Return request {id} not found.");

        // Guard invalid transitions
        if (returnRequest.Status == ReturnStatus.Completed)
            throw new InvalidOperationException("A completed return request cannot be updated.");

        if (returnRequest.Status == ReturnStatus.Rejected && status != ReturnStatus.Pending)
            throw new InvalidOperationException("A rejected return can only be re-opened to Pending.");

        returnRequest.Status = status;
        returnRequest.UpdatedAt = DateTime.UtcNow;

        await _db.SaveChangesAsync();
        return returnRequest;
    }

    public async Task<RefundRecord> ProcessRefundAsync(Guid returnId, ProcessRefundRequest request)
    {
        if (request.Amount <= 0)
            throw new ArgumentException("Refund amount must be greater than zero.", nameof(request));

        var validMethods = new[] { "original", "store_credit" };
        if (!validMethods.Contains(request.Method))
            throw new ArgumentException($"Method must be one of: {string.Join(", ", validMethods)}.", nameof(request));

        _logger.LogInformation(
            "Processing refund of {Amount} {Currency} ({Method}) for return {ReturnId}",
            request.Amount, request.Currency, request.Method, returnId);

        var returnRequest = await _db.ReturnRequests
            .Include(r => r.Refund)
            .FirstOrDefaultAsync(r => r.Id == returnId)
            ?? throw new KeyNotFoundException($"Return request {returnId} not found.");

        if (returnRequest.Status != ReturnStatus.Approved)
            throw new InvalidOperationException(
                $"Refunds can only be processed for Approved returns. Current status: {returnRequest.Status}.");

        if (returnRequest.Refund is not null)
            throw new InvalidOperationException($"A refund has already been processed for return {returnId}.");

        var refund = new RefundRecord
        {
            Id = Guid.NewGuid(),
            ReturnRequestId = returnId,
            Amount = request.Amount,
            Currency = request.Currency.ToUpperInvariant(),
            Method = request.Method,
            ProcessedAt = DateTime.UtcNow
        };

        _db.RefundRecords.Add(refund);

        // Automatically mark the return as Completed once refund is recorded
        returnRequest.Status = ReturnStatus.Completed;
        returnRequest.UpdatedAt = DateTime.UtcNow;

        await _db.SaveChangesAsync();
        return refund;
    }
}
