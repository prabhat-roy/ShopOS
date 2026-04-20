using ReturnRefundService.Models;

namespace ReturnRefundService.Services;

public interface IReturnService
{
    /// <summary>Get a single return request by its ID.</summary>
    Task<ReturnRequest?> GetReturnAsync(Guid id);

    /// <summary>List all return requests for a given customer.</summary>
    Task<IReadOnlyList<ReturnRequest>> ListReturnsAsync(string customerId);

    /// <summary>Create a new return request (RMA).</summary>
    Task<ReturnRequest> CreateReturnAsync(CreateReturnRequest request);

    /// <summary>Update the status of an existing return request.</summary>
    Task<ReturnRequest> UpdateStatusAsync(Guid id, ReturnStatus status);

    /// <summary>Record a refund against an approved return request.</summary>
    Task<RefundRecord> ProcessRefundAsync(Guid returnId, ProcessRefundRequest request);
}
