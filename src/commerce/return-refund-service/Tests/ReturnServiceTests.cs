using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging.Abstractions;
using ReturnRefundService.Data;
using ReturnRefundService.Models;
using ReturnRefundService.Services;
using Xunit;

namespace ReturnRefundService.Tests;

/// <summary>
/// Unit tests for ReturnServiceImpl using EF Core InMemory provider.
/// Each test creates its own isolated DbContext to avoid state leakage.
/// </summary>
public class ReturnServiceTests
{
    private static AppDbContext CreateContext(string dbName)
    {
        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseInMemoryDatabase(dbName)
            .Options;
        return new AppDbContext(options);
    }

    private static ReturnServiceImpl CreateService(AppDbContext ctx)
        => new(ctx, NullLogger<ReturnServiceImpl>.Instance);

    // Helper: persist a return in a given status
    private static async Task<ReturnRequest> SeedReturn(
        AppDbContext ctx,
        ReturnStatus status = ReturnStatus.Pending)
    {
        var rr = new ReturnRequest
        {
            Id = Guid.NewGuid(),
            OrderId = "order-001",
            CustomerId = "customer-abc",
            ProductId = "prod-xyz",
            Quantity = 1,
            Reason = ReturnReason.Defective,
            Notes = "Item arrived broken.",
            Status = status,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };
        ctx.ReturnRequests.Add(rr);
        await ctx.SaveChangesAsync();
        return rr;
    }

    // ── CreateReturn ──────────────────────────────────────────────────────

    [Fact]
    public async Task CreateReturn_ValidRequest_PersistsAndReturnsPending()
    {
        await using var ctx = CreateContext(nameof(CreateReturn_ValidRequest_PersistsAndReturnsPending));
        var svc = CreateService(ctx);

        var req = new CreateReturnRequest("order-001", "customer-abc", "prod-xyz", 2, ReturnReason.Defective, "Broken.");
        var result = await svc.CreateReturnAsync(req);

        Assert.NotEqual(Guid.Empty, result.Id);
        Assert.Equal(ReturnStatus.Pending, result.Status);
        Assert.Equal("order-001", result.OrderId);
        Assert.Equal("customer-abc", result.CustomerId);
        Assert.Equal("prod-xyz", result.ProductId);
        Assert.Equal(2, result.Quantity);
        Assert.Equal(ReturnReason.Defective, result.Reason);
        Assert.Equal("Broken.", result.Notes);

        // Verify it was persisted in the database
        var persisted = await ctx.ReturnRequests.FindAsync(result.Id);
        Assert.NotNull(persisted);
    }

    [Fact]
    public async Task CreateReturn_MissingOrderId_ThrowsArgumentException()
    {
        await using var ctx = CreateContext(nameof(CreateReturn_MissingOrderId_ThrowsArgumentException));
        var svc = CreateService(ctx);

        var req = new CreateReturnRequest("", "customer-abc", "prod-xyz", 1, ReturnReason.Other);
        await Assert.ThrowsAsync<ArgumentException>(() => svc.CreateReturnAsync(req));
    }

    [Fact]
    public async Task CreateReturn_ZeroQuantity_ThrowsArgumentException()
    {
        await using var ctx = CreateContext(nameof(CreateReturn_ZeroQuantity_ThrowsArgumentException));
        var svc = CreateService(ctx);

        var req = new CreateReturnRequest("order-001", "customer-abc", "prod-xyz", 0, ReturnReason.ChangedMind);
        await Assert.ThrowsAsync<ArgumentException>(() => svc.CreateReturnAsync(req));
    }

    // ── GetReturn ─────────────────────────────────────────────────────────

    [Fact]
    public async Task GetReturn_ExistingId_ReturnsRequest()
    {
        await using var ctx = CreateContext(nameof(GetReturn_ExistingId_ReturnsRequest));
        var seeded = await SeedReturn(ctx);
        var svc = CreateService(ctx);

        var result = await svc.GetReturnAsync(seeded.Id);

        Assert.NotNull(result);
        Assert.Equal(seeded.Id, result!.Id);
    }

    [Fact]
    public async Task GetReturn_NonExistingId_ReturnsNull()
    {
        await using var ctx = CreateContext(nameof(GetReturn_NonExistingId_ReturnsNull));
        var svc = CreateService(ctx);

        var result = await svc.GetReturnAsync(Guid.NewGuid());

        Assert.Null(result);
    }

    // ── ListReturns ───────────────────────────────────────────────────────

    [Fact]
    public async Task ListReturns_FiltersToCustomer()
    {
        await using var ctx = CreateContext(nameof(ListReturns_FiltersToCustomer));

        // Seed two returns for customer-A, one for customer-B
        ctx.ReturnRequests.AddRange(
            new ReturnRequest { Id = Guid.NewGuid(), OrderId = "o1", CustomerId = "A", ProductId = "p1", Quantity = 1, Reason = ReturnReason.Other, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow },
            new ReturnRequest { Id = Guid.NewGuid(), OrderId = "o2", CustomerId = "A", ProductId = "p2", Quantity = 2, Reason = ReturnReason.Defective, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow },
            new ReturnRequest { Id = Guid.NewGuid(), OrderId = "o3", CustomerId = "B", ProductId = "p3", Quantity = 1, Reason = ReturnReason.WrongItem, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow }
        );
        await ctx.SaveChangesAsync();

        var svc = CreateService(ctx);
        var results = await svc.ListReturnsAsync("A");

        Assert.Equal(2, results.Count);
        Assert.All(results, r => Assert.Equal("A", r.CustomerId));
    }

    [Fact]
    public async Task ListReturns_EmptyCustomerId_ThrowsArgumentException()
    {
        await using var ctx = CreateContext(nameof(ListReturns_EmptyCustomerId_ThrowsArgumentException));
        var svc = CreateService(ctx);

        await Assert.ThrowsAsync<ArgumentException>(() => svc.ListReturnsAsync(""));
    }

    // ── UpdateStatus ──────────────────────────────────────────────────────

    [Fact]
    public async Task UpdateStatus_PendingToApproved_UpdatesSuccessfully()
    {
        await using var ctx = CreateContext(nameof(UpdateStatus_PendingToApproved_UpdatesSuccessfully));
        var seeded = await SeedReturn(ctx, ReturnStatus.Pending);
        var svc = CreateService(ctx);

        var result = await svc.UpdateStatusAsync(seeded.Id, ReturnStatus.Approved);

        Assert.Equal(ReturnStatus.Approved, result.Status);
        Assert.True(result.UpdatedAt >= seeded.UpdatedAt);
    }

    [Fact]
    public async Task UpdateStatus_PendingToRejected_UpdatesSuccessfully()
    {
        await using var ctx = CreateContext(nameof(UpdateStatus_PendingToRejected_UpdatesSuccessfully));
        var seeded = await SeedReturn(ctx, ReturnStatus.Pending);
        var svc = CreateService(ctx);

        var result = await svc.UpdateStatusAsync(seeded.Id, ReturnStatus.Rejected);

        Assert.Equal(ReturnStatus.Rejected, result.Status);
    }

    [Fact]
    public async Task UpdateStatus_CompletedReturn_ThrowsInvalidOperationException()
    {
        await using var ctx = CreateContext(nameof(UpdateStatus_CompletedReturn_ThrowsInvalidOperationException));
        var seeded = await SeedReturn(ctx, ReturnStatus.Completed);
        var svc = CreateService(ctx);

        await Assert.ThrowsAsync<InvalidOperationException>(() =>
            svc.UpdateStatusAsync(seeded.Id, ReturnStatus.Approved));
    }

    [Fact]
    public async Task UpdateStatus_NonExistingId_ThrowsKeyNotFoundException()
    {
        await using var ctx = CreateContext(nameof(UpdateStatus_NonExistingId_ThrowsKeyNotFoundException));
        var svc = CreateService(ctx);

        await Assert.ThrowsAsync<KeyNotFoundException>(() =>
            svc.UpdateStatusAsync(Guid.NewGuid(), ReturnStatus.Approved));
    }

    // ── ProcessRefund ─────────────────────────────────────────────────────

    [Fact]
    public async Task ProcessRefund_ApprovedReturn_CreatesRefundAndCompletesReturn()
    {
        await using var ctx = CreateContext(nameof(ProcessRefund_ApprovedReturn_CreatesRefundAndCompletesReturn));
        var seeded = await SeedReturn(ctx, ReturnStatus.Approved);
        var svc = CreateService(ctx);

        var refund = await svc.ProcessRefundAsync(seeded.Id, new ProcessRefundRequest(49.99m, "original", "USD"));

        Assert.NotEqual(Guid.Empty, refund.Id);
        Assert.Equal(seeded.Id, refund.ReturnRequestId);
        Assert.Equal(49.99m, refund.Amount);
        Assert.Equal("USD", refund.Currency);
        Assert.Equal("original", refund.Method);

        // Return should now be Completed
        var updated = await ctx.ReturnRequests.FindAsync(seeded.Id);
        Assert.Equal(ReturnStatus.Completed, updated!.Status);
    }

    [Fact]
    public async Task ProcessRefund_StoreCredit_RecordsCorrectMethod()
    {
        await using var ctx = CreateContext(nameof(ProcessRefund_StoreCredit_RecordsCorrectMethod));
        var seeded = await SeedReturn(ctx, ReturnStatus.Approved);
        var svc = CreateService(ctx);

        var refund = await svc.ProcessRefundAsync(seeded.Id, new ProcessRefundRequest(20.00m, "store_credit", "USD"));

        Assert.Equal("store_credit", refund.Method);
    }

    [Fact]
    public async Task ProcessRefund_PendingReturn_ThrowsInvalidOperationException()
    {
        await using var ctx = CreateContext(nameof(ProcessRefund_PendingReturn_ThrowsInvalidOperationException));
        var seeded = await SeedReturn(ctx, ReturnStatus.Pending);
        var svc = CreateService(ctx);

        await Assert.ThrowsAsync<InvalidOperationException>(() =>
            svc.ProcessRefundAsync(seeded.Id, new ProcessRefundRequest(10.00m)));
    }

    [Fact]
    public async Task ProcessRefund_DuplicateRefund_ThrowsInvalidOperationException()
    {
        await using var ctx = CreateContext(nameof(ProcessRefund_DuplicateRefund_ThrowsInvalidOperationException));
        var seeded = await SeedReturn(ctx, ReturnStatus.Approved);
        var svc = CreateService(ctx);

        // First refund succeeds
        await svc.ProcessRefundAsync(seeded.Id, new ProcessRefundRequest(10.00m));

        // The return is now Completed; second attempt should fail
        await Assert.ThrowsAsync<InvalidOperationException>(() =>
            svc.ProcessRefundAsync(seeded.Id, new ProcessRefundRequest(10.00m)));
    }

    [Fact]
    public async Task ProcessRefund_ZeroAmount_ThrowsArgumentException()
    {
        await using var ctx = CreateContext(nameof(ProcessRefund_ZeroAmount_ThrowsArgumentException));
        var seeded = await SeedReturn(ctx, ReturnStatus.Approved);
        var svc = CreateService(ctx);

        await Assert.ThrowsAsync<ArgumentException>(() =>
            svc.ProcessRefundAsync(seeded.Id, new ProcessRefundRequest(0m)));
    }

    [Fact]
    public async Task ProcessRefund_InvalidMethod_ThrowsArgumentException()
    {
        await using var ctx = CreateContext(nameof(ProcessRefund_InvalidMethod_ThrowsArgumentException));
        var seeded = await SeedReturn(ctx, ReturnStatus.Approved);
        var svc = CreateService(ctx);

        await Assert.ThrowsAsync<ArgumentException>(() =>
            svc.ProcessRefundAsync(seeded.Id, new ProcessRefundRequest(10.00m, "cash")));
    }

    [Fact]
    public async Task ProcessRefund_NonExistingReturn_ThrowsKeyNotFoundException()
    {
        await using var ctx = CreateContext(nameof(ProcessRefund_NonExistingReturn_ThrowsKeyNotFoundException));
        var svc = CreateService(ctx);

        await Assert.ThrowsAsync<KeyNotFoundException>(() =>
            svc.ProcessRefundAsync(Guid.NewGuid(), new ProcessRefundRequest(10.00m)));
    }
}
