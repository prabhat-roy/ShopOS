using System.Text.Json;
using CartService.Models;

namespace CartService.Services;

public class CartServiceImpl : ICartService
{
    private readonly IRedisService _redis;
    private readonly IConfiguration _configuration;
    private readonly ILogger<CartServiceImpl> _logger;
    private readonly TimeSpan _cartTtl;

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true,
        WriteIndented = false
    };

    public CartServiceImpl(IRedisService redis, IConfiguration configuration, ILogger<CartServiceImpl> logger)
    {
        _redis = redis;
        _configuration = configuration;
        _logger = logger;

        var ttlHours = int.TryParse(configuration["CART_TTL_HOURS"], out var h) ? h : 48;
        _cartTtl = TimeSpan.FromHours(ttlHours);
    }

    private static string CartKey(string userId) => $"cart:{userId}";

    private async Task<Cart> LoadCartAsync(string userId)
    {
        var json = await _redis.GetAsync(CartKey(userId));
        if (json is null)
        {
            return new Cart
            {
                UserId = userId,
                UpdatedAt = DateTime.UtcNow
            };
        }

        var cart = JsonSerializer.Deserialize<Cart>(json, JsonOptions);
        return cart ?? new Cart { UserId = userId, UpdatedAt = DateTime.UtcNow };
    }

    private async Task<Cart> SaveCartAsync(Cart cart)
    {
        cart.UpdatedAt = DateTime.UtcNow;
        var json = JsonSerializer.Serialize(cart, JsonOptions);
        await _redis.SetAsync(CartKey(cart.UserId), json, _cartTtl);
        return cart;
    }

    public async Task<Cart> GetCartAsync(string userId)
    {
        _logger.LogDebug("Getting cart for user {UserId}", userId);
        return await LoadCartAsync(userId);
    }

    public async Task<Cart> AddItemAsync(string userId, AddItemRequest request)
    {
        if (string.IsNullOrWhiteSpace(request.ProductId))
            throw new ArgumentException("ProductId is required.", nameof(request));
        if (request.Quantity <= 0)
            throw new ArgumentException("Quantity must be greater than zero.", nameof(request));
        if (request.Price < 0)
            throw new ArgumentException("Price cannot be negative.", nameof(request));

        _logger.LogDebug("Adding item {ProductId} to cart for user {UserId}", request.ProductId, userId);

        var cart = await LoadCartAsync(userId);

        var existing = cart.Items.FirstOrDefault(i => i.ProductId == request.ProductId);
        if (existing is not null)
        {
            existing.Quantity += request.Quantity;
            existing.Price = request.Price;
            existing.Name = request.Name;
            existing.SKU = request.SKU;
            existing.ImageUrl = request.ImageUrl;
        }
        else
        {
            cart.Items.Add(new CartItem
            {
                ProductId = request.ProductId,
                SKU = request.SKU,
                Name = request.Name,
                Price = request.Price,
                Quantity = request.Quantity,
                ImageUrl = request.ImageUrl
            });
        }

        return await SaveCartAsync(cart);
    }

    public async Task<Cart> UpdateQuantityAsync(string userId, string productId, int quantity)
    {
        _logger.LogDebug("Updating quantity of {ProductId} to {Quantity} for user {UserId}", productId, quantity, userId);

        var cart = await LoadCartAsync(userId);

        if (quantity <= 0)
        {
            cart.Items.RemoveAll(i => i.ProductId == productId);
        }
        else
        {
            var item = cart.Items.FirstOrDefault(i => i.ProductId == productId);
            if (item is not null)
                item.Quantity = quantity;
        }

        return await SaveCartAsync(cart);
    }

    public async Task<Cart> RemoveItemAsync(string userId, string productId)
    {
        _logger.LogDebug("Removing item {ProductId} from cart for user {UserId}", productId, userId);

        var cart = await LoadCartAsync(userId);
        cart.Items.RemoveAll(i => i.ProductId == productId);
        return await SaveCartAsync(cart);
    }

    public async Task<Cart> ApplyCouponAsync(string userId, ApplyCouponRequest request)
    {
        if (string.IsNullOrWhiteSpace(request.Code))
            throw new ArgumentException("Coupon code is required.", nameof(request));
        if (request.DiscountAmount < 0)
            throw new ArgumentException("Discount amount cannot be negative.", nameof(request));

        _logger.LogDebug("Applying coupon {Code} to cart for user {UserId}", request.Code, userId);

        var cart = await LoadCartAsync(userId);
        cart.CouponCode = request.Code.Trim().ToUpperInvariant();
        cart.Discount = request.DiscountAmount;
        return await SaveCartAsync(cart);
    }

    public async Task ClearCartAsync(string userId)
    {
        _logger.LogDebug("Clearing cart for user {UserId}", userId);
        await _redis.DeleteAsync(CartKey(userId));
    }

    public async Task<CartSummary> GetCartSummaryAsync(string userId)
    {
        var cart = await LoadCartAsync(userId);
        var itemCount = cart.Items.Sum(i => i.Quantity);
        return new CartSummary(itemCount, cart.Total);
    }
}
