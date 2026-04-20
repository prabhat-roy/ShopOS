using Microsoft.AspNetCore.Mvc;
using ReturnRefundService.Models;
using ReturnRefundService.Services;

namespace ReturnRefundService.Controllers;

[ApiController]
[Route("returns")]
[Produces("application/json")]
public class ReturnController : ControllerBase
{
    private readonly IReturnService _returnService;
    private readonly ILogger<ReturnController> _logger;

    public ReturnController(IReturnService returnService, ILogger<ReturnController> logger)
    {
        _returnService = returnService;
        _logger = logger;
    }

    /// <summary>Create a new return request (RMA).</summary>
    [HttpPost]
    [ProducesResponseType(typeof(ReturnRequest), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> CreateReturn([FromBody] CreateReturnRequest request)
    {
        if (!ModelState.IsValid)
            return BadRequest(ModelState);

        try
        {
            var returnRequest = await _returnService.CreateReturnAsync(request);
            return CreatedAtAction(nameof(GetReturn), new { id = returnRequest.Id }, returnRequest);
        }
        catch (ArgumentException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    /// <summary>Get a return request by ID.</summary>
    [HttpGet("{id:guid}")]
    [ProducesResponseType(typeof(ReturnRequest), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetReturn([FromRoute] Guid id)
    {
        var returnRequest = await _returnService.GetReturnAsync(id);
        if (returnRequest is null)
            return NotFound(new { error = $"Return request {id} not found." });

        return Ok(returnRequest);
    }

    /// <summary>List return requests, filtered by customerId query parameter.</summary>
    [HttpGet]
    [ProducesResponseType(typeof(IReadOnlyList<ReturnRequest>), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> ListReturns([FromQuery] string? customerId)
    {
        if (string.IsNullOrWhiteSpace(customerId))
            return BadRequest(new { error = "Query parameter 'customerId' is required." });

        try
        {
            var returns = await _returnService.ListReturnsAsync(customerId);
            return Ok(returns);
        }
        catch (ArgumentException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    /// <summary>Update the status of a return request.</summary>
    [HttpPatch("{id:guid}/status")]
    [ProducesResponseType(typeof(ReturnRequest), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> UpdateStatus(
        [FromRoute] Guid id,
        [FromBody] UpdateStatusRequest request)
    {
        if (!ModelState.IsValid)
            return BadRequest(ModelState);

        try
        {
            var returnRequest = await _returnService.UpdateStatusAsync(id, request.Status);
            return Ok(returnRequest);
        }
        catch (KeyNotFoundException ex)
        {
            return NotFound(new { error = ex.Message });
        }
        catch (InvalidOperationException ex)
        {
            return Conflict(new { error = ex.Message });
        }
    }

    /// <summary>Process a refund for an approved return.</summary>
    [HttpPost("{id:guid}/refund")]
    [ProducesResponseType(typeof(RefundRecord), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> ProcessRefund(
        [FromRoute] Guid id,
        [FromBody] ProcessRefundRequest request)
    {
        if (!ModelState.IsValid)
            return BadRequest(ModelState);

        try
        {
            var refund = await _returnService.ProcessRefundAsync(id, request);
            return CreatedAtAction(nameof(GetReturn), new { id }, refund);
        }
        catch (ArgumentException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
        catch (KeyNotFoundException ex)
        {
            return NotFound(new { error = ex.Message });
        }
        catch (InvalidOperationException ex)
        {
            return Conflict(new { error = ex.Message });
        }
    }
}
