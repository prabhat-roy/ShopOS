using Microsoft.AspNetCore.Mvc;
using CartService.Models;
using CartService.Services;

namespace CartService.Controllers;

[ApiController]
[Route("carts")]
[Produces("application/json")]
public class CartController : ControllerBase
{
    private readonly ICartService _cartService;
    private readonly ILogger<CartController> _logger;

    public CartController(ICartService cartService, ILogger<CartController> logger)
    {
        _cartService = cartService;
        _logger = logger;
    }

    /// <summary>Get the full cart for a user.</summary>
    [HttpGet("{userId}")]
    [ProducesResponseType(typeof(Cart), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetCart([FromRoute] string userId)
    {
        var cart = await _cartService.GetCartAsync(userId);
        return Ok(cart);
    }

    /// <summary>Add or upsert an item in the cart.</summary>
    [HttpPost("{userId}/items")]
    [ProducesResponseType(typeof(Cart), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> AddItem([FromRoute] string userId, [FromBody] AddItemRequest request)
    {
        if (!ModelState.IsValid)
            return BadRequest(ModelState);

        try
        {
            var cart = await _cartService.AddItemAsync(userId, request);
            return CreatedAtAction(nameof(GetCart), new { userId }, cart);
        }
        catch (ArgumentException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    /// <summary>Update the quantity of a specific item. Set quantity to 0 to remove.</summary>
    [HttpPut("{userId}/items/{productId}")]
    [ProducesResponseType(typeof(Cart), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> UpdateQuantity(
        [FromRoute] string userId,
        [FromRoute] string productId,
        [FromBody] UpdateQuantityRequest request)
    {
        if (!ModelState.IsValid)
            return BadRequest(ModelState);

        var cart = await _cartService.UpdateQuantityAsync(userId, productId, request.Quantity);
        return Ok(cart);
    }

    /// <summary>Remove a specific item from the cart.</summary>
    [HttpDelete("{userId}/items/{productId}")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    public async Task<IActionResult> RemoveItem([FromRoute] string userId, [FromRoute] string productId)
    {
        await _cartService.RemoveItemAsync(userId, productId);
        return NoContent();
    }

    /// <summary>Apply a coupon code to the cart.</summary>
    [HttpPost("{userId}/coupon")]
    [ProducesResponseType(typeof(Cart), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> ApplyCoupon(
        [FromRoute] string userId,
        [FromBody] ApplyCouponRequest request)
    {
        if (!ModelState.IsValid)
            return BadRequest(ModelState);

        try
        {
            var cart = await _cartService.ApplyCouponAsync(userId, request);
            return Ok(cart);
        }
        catch (ArgumentException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    /// <summary>Clear (delete) the entire cart for a user.</summary>
    [HttpDelete("{userId}")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    public async Task<IActionResult> ClearCart([FromRoute] string userId)
    {
        await _cartService.ClearCartAsync(userId);
        return NoContent();
    }

    /// <summary>Get a lightweight summary of the cart (item count + total).</summary>
    [HttpGet("{userId}/summary")]
    [ProducesResponseType(typeof(CartSummary), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetCartSummary([FromRoute] string userId)
    {
        var summary = await _cartService.GetCartSummaryAsync(userId);
        return Ok(summary);
    }
}
