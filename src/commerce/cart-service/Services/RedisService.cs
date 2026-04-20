using StackExchange.Redis;

namespace CartService.Services;

public class RedisService : IRedisService
{
    private readonly IDatabase _db;
    private readonly ILogger<RedisService> _logger;

    public RedisService(IConnectionMultiplexer redis, IConfiguration configuration, ILogger<RedisService> logger)
    {
        _logger = logger;
        var dbIndex = int.TryParse(configuration["REDIS_DB"], out var idx) ? idx : 0;
        _db = redis.GetDatabase(dbIndex);
    }

    public async Task<string?> GetAsync(string key)
    {
        try
        {
            var value = await _db.StringGetAsync(key);
            return value.HasValue ? value.ToString() : null;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Redis GET failed for key {Key}", key);
            throw;
        }
    }

    public async Task SetAsync(string key, string value, TimeSpan? ttl = null)
    {
        try
        {
            await _db.StringSetAsync(key, value, ttl);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Redis SET failed for key {Key}", key);
            throw;
        }
    }

    public async Task DeleteAsync(string key)
    {
        try
        {
            await _db.KeyDeleteAsync(key);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Redis DELETE failed for key {Key}", key);
            throw;
        }
    }

    public async Task<bool> ExistsAsync(string key)
    {
        try
        {
            return await _db.KeyExistsAsync(key);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Redis EXISTS failed for key {Key}", key);
            throw;
        }
    }
}
