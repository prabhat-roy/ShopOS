using CartService.Models;

namespace CartService.Services;

public interface ICartService
{
    Task<Cart> GetCartAsync(string userId);
    Task<Cart> AddItemAsync(string userId, AddItemRequest request);
    Task<Cart> UpdateQuantityAsync(string userId, string productId, int quantity);
    Task<Cart> RemoveItemAsync(string userId, string productId);
    Task<Cart> ApplyCouponAsync(string userId, ApplyCouponRequest request);
    Task ClearCartAsync(string userId);
    Task<CartSummary> GetCartSummaryAsync(string userId);
}
