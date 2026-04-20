namespace CartService.Services;

public interface IRedisService
{
    Task<string?> GetAsync(string key);
    Task SetAsync(string key, string value, TimeSpan? ttl = null);
    Task DeleteAsync(string key);
    Task<bool> ExistsAsync(string key);
}
