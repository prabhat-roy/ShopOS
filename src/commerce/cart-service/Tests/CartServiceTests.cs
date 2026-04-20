using System.Text.Json;
using CartService.Models;
using CartService.Services;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging.Abstractions;
using Moq;
using Xunit;

namespace CartService.Tests;

public class CartServiceTests
{
    private readonly Mock<IRedisService> _redisMock;
    private readonly IConfiguration _configuration;
    private readonly CartServiceImpl _sut;

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public CartServiceTests()
    {
        _redisMock = new Mock<IRedisService>();

        var inMemory = new Dictionary<string, string?>
        {
            ["CART_TTL_HOURS"] = "48",
            ["REDIS_DB"] = "0"
        };
        _configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(inMemory)
            .Build();

        _sut = new CartServiceImpl(_redisMock.Object, _configuration, NullLogger<CartServiceImpl>.Instance);
    }

    // Helper: simulate an empty Redis (cart does not exist yet)
    private void SetupEmptyCart() =>
        _redisMock.Setup(r => r.GetAsync(It.IsAny<string>())).ReturnsAsync((string?)null);

    // Helper: simulate a pre-existing cart stored in Redis
    private void SetupExistingCart(Cart cart)
    {
        var json = JsonSerializer.Serialize(cart);
        _redisMock.Setup(r => r.GetAsync(It.IsAny<string>())).ReturnsAsync(json);
    }

    // Capture what was written to Redis via SetAsync
    private string? CaptureSetAsync()
    {
        string? captured = null;
        _redisMock
            .Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
            .Callback<string, string, TimeSpan?>((_, v, _) => captured = v)
            .Returns(Task.CompletedTask);
        return captured; // Will be populated after the act step
    }

    // ── AddItem tests ────────────────────────────────────────────────────────

    [Fact]
    public async Task AddItem_EmptyCart_AddsNewItem()
    {
        SetupEmptyCart();
        string? savedJson = null;
        _redisMock
            .Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
            .Callback<string, string, TimeSpan?>((_, v, _) => savedJson = v)
            .Returns(Task.CompletedTask);

        var req = new AddItemRequest("prod-1", "SKU-001", "Widget", 9.99m, 2, "https://img.example.com/1.jpg");
        var cart = await _sut.AddItemAsync("user-1", req);

        Assert.Single(cart.Items);
        Assert.Equal("prod-1", cart.Items[0].ProductId);
        Assert.Equal(2, cart.Items[0].Quantity);
        Assert.Equal(9.99m, cart.Items[0].Price);
        Assert.NotNull(savedJson);
    }

    [Fact]
    public async Task AddItem_ExistingProduct_IncrementsQuantity()
    {
        var existing = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "prod-1", SKU = "SKU-001", Name = "Widget", Price = 9.99m, Quantity = 1 }
            }
        };
        SetupExistingCart(existing);
        _redisMock.Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
                  .Returns(Task.CompletedTask);

        var req = new AddItemRequest("prod-1", "SKU-001", "Widget", 9.99m, 3);
        var cart = await _sut.AddItemAsync("user-1", req);

        Assert.Single(cart.Items);
        Assert.Equal(4, cart.Items[0].Quantity); // 1 existing + 3 new
    }

    [Fact]
    public async Task AddItem_MultipleProducts_AddsDistinctItems()
    {
        SetupEmptyCart();
        _redisMock.Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
                  .Returns(Task.CompletedTask);

        // We need a cart that already has prod-1 for the second call
        var cartWithOne = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "prod-1", SKU = "SKU-001", Name = "Widget", Price = 9.99m, Quantity = 1 }
            }
        };

        // First add returns empty, second add returns cart-with-one
        _redisMock
            .SetupSequence(r => r.GetAsync(It.IsAny<string>()))
            .ReturnsAsync((string?)null)
            .ReturnsAsync(JsonSerializer.Serialize(cartWithOne));

        await _sut.AddItemAsync("user-1", new AddItemRequest("prod-1", "SKU-001", "Widget", 9.99m, 1));
        var cart = await _sut.AddItemAsync("user-1", new AddItemRequest("prod-2", "SKU-002", "Gadget", 19.99m, 1));

        Assert.Equal(2, cart.Items.Count);
    }

    [Fact]
    public async Task AddItem_InvalidProductId_ThrowsArgumentException()
    {
        SetupEmptyCart();
        var req = new AddItemRequest("", "SKU-001", "Widget", 9.99m, 1);
        await Assert.ThrowsAsync<ArgumentException>(() => _sut.AddItemAsync("user-1", req));
    }

    [Fact]
    public async Task AddItem_ZeroQuantity_ThrowsArgumentException()
    {
        SetupEmptyCart();
        var req = new AddItemRequest("prod-1", "SKU-001", "Widget", 9.99m, 0);
        await Assert.ThrowsAsync<ArgumentException>(() => _sut.AddItemAsync("user-1", req));
    }

    // ── UpdateQuantity tests ─────────────────────────────────────────────────

    [Fact]
    public async Task UpdateQuantity_ZeroQuantity_RemovesItem()
    {
        var existing = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "prod-1", SKU = "SKU-001", Name = "Widget", Price = 9.99m, Quantity = 2 }
            }
        };
        SetupExistingCart(existing);
        _redisMock.Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
                  .Returns(Task.CompletedTask);

        var cart = await _sut.UpdateQuantityAsync("user-1", "prod-1", 0);

        Assert.Empty(cart.Items);
    }

    [Fact]
    public async Task UpdateQuantity_NegativeQuantity_RemovesItem()
    {
        var existing = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "prod-1", SKU = "SKU-001", Name = "Widget", Price = 9.99m, Quantity = 3 }
            }
        };
        SetupExistingCart(existing);
        _redisMock.Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
                  .Returns(Task.CompletedTask);

        var cart = await _sut.UpdateQuantityAsync("user-1", "prod-1", -1);

        Assert.Empty(cart.Items);
    }

    [Fact]
    public async Task UpdateQuantity_PositiveQuantity_UpdatesItem()
    {
        var existing = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "prod-1", SKU = "SKU-001", Name = "Widget", Price = 9.99m, Quantity = 2 }
            }
        };
        SetupExistingCart(existing);
        _redisMock.Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
                  .Returns(Task.CompletedTask);

        var cart = await _sut.UpdateQuantityAsync("user-1", "prod-1", 5);

        Assert.Single(cart.Items);
        Assert.Equal(5, cart.Items[0].Quantity);
    }

    // ── ClearCart tests ──────────────────────────────────────────────────────

    [Fact]
    public async Task ClearCart_CallsRedisDelete()
    {
        _redisMock.Setup(r => r.DeleteAsync(It.IsAny<string>())).Returns(Task.CompletedTask);

        await _sut.ClearCartAsync("user-1");

        _redisMock.Verify(r => r.DeleteAsync("cart:user-1"), Times.Once);
    }

    // ── ApplyCoupon tests ────────────────────────────────────────────────────

    [Fact]
    public async Task ApplyCoupon_ValidCoupon_SetsCouponAndDiscount()
    {
        SetupEmptyCart();
        string? savedJson = null;
        _redisMock
            .Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
            .Callback<string, string, TimeSpan?>((_, v, _) => savedJson = v)
            .Returns(Task.CompletedTask);

        await _sut.ApplyCouponAsync("user-1", new ApplyCouponRequest("SAVE10", 10.00m));

        Assert.NotNull(savedJson);
        var saved = JsonSerializer.Deserialize<Cart>(savedJson!, JsonOptions);
        Assert.Equal("SAVE10", saved!.CouponCode);
        Assert.Equal(10.00m, saved.Discount);
    }

    [Fact]
    public async Task ApplyCoupon_NormalizesCodeToUppercase()
    {
        SetupEmptyCart();
        string? savedJson = null;
        _redisMock
            .Setup(r => r.SetAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<TimeSpan?>()))
            .Callback<string, string, TimeSpan?>((_, v, _) => savedJson = v)
            .Returns(Task.CompletedTask);

        await _sut.ApplyCouponAsync("user-1", new ApplyCouponRequest("summer25", 25.00m));

        var saved = JsonSerializer.Deserialize<Cart>(savedJson!, JsonOptions);
        Assert.Equal("SUMMER25", saved!.CouponCode);
    }

    [Fact]
    public async Task ApplyCoupon_EmptyCode_ThrowsArgumentException()
    {
        SetupEmptyCart();
        await Assert.ThrowsAsync<ArgumentException>(() =>
            _sut.ApplyCouponAsync("user-1", new ApplyCouponRequest("", 5.00m)));
    }

    // ── Computed properties ──────────────────────────────────────────────────

    [Fact]
    public async Task GetCart_SubtotalAndTotalCalculatedCorrectly()
    {
        var existing = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "p1", Price = 10.00m, Quantity = 2 },  // 20.00
                new() { ProductId = "p2", Price = 5.50m,  Quantity = 1 }   //  5.50
            },
            Discount = 3.00m
        };
        SetupExistingCart(existing);

        var cart = await _sut.GetCartAsync("user-1");

        Assert.Equal(25.50m, cart.Subtotal);
        Assert.Equal(22.50m, cart.Total);   // 25.50 - 3.00
    }

    [Fact]
    public async Task GetCart_TotalNeverNegative()
    {
        var existing = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "p1", Price = 5.00m, Quantity = 1 }
            },
            Discount = 100.00m  // discount larger than subtotal
        };
        SetupExistingCart(existing);

        var cart = await _sut.GetCartAsync("user-1");

        Assert.Equal(0m, cart.Total);
    }

    // ── GetCartSummary tests ─────────────────────────────────────────────────

    [Fact]
    public async Task GetCartSummary_ReturnsCorrectItemCountAndTotal()
    {
        var existing = new Cart
        {
            UserId = "user-1",
            Items = new List<CartItem>
            {
                new() { ProductId = "p1", Price = 10.00m, Quantity = 3 },
                new() { ProductId = "p2", Price = 20.00m, Quantity = 1 }
            },
            Discount = 0m
        };
        SetupExistingCart(existing);

        var summary = await _sut.GetCartSummaryAsync("user-1");

        Assert.Equal(4, summary.ItemCount);  // 3 + 1
        Assert.Equal(50.00m, summary.Total); // (10*3) + (20*1)
    }
}
