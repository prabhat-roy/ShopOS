namespace CartService.Models;

public class Cart
{
    public string UserId { get; set; } = "";
    public List<CartItem> Items { get; set; } = new();
    public string? CouponCode { get; set; }
    public decimal Discount { get; set; }
    public decimal Subtotal => Items.Sum(i => i.Price * i.Quantity);
    public decimal Total => Math.Max(0, Subtotal - Discount);
    public DateTime UpdatedAt { get; set; }
}

public class CartItem
{
    public string ProductId { get; set; } = "";
    public string SKU { get; set; } = "";
    public string Name { get; set; } = "";
    public decimal Price { get; set; }
    public int Quantity { get; set; }
    public string ImageUrl { get; set; } = "";
}

public record AddItemRequest(
    string ProductId,
    string SKU,
    string Name,
    decimal Price,
    int Quantity,
    string ImageUrl = "");

public record ApplyCouponRequest(string Code, decimal DiscountAmount);

public record UpdateQuantityRequest(int Quantity);

public record CartSummary(int ItemCount, decimal Total);
