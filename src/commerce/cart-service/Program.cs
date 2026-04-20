using StackExchange.Redis;
using CartService.Services;

var builder = WebApplication.CreateBuilder(args);

// ── Configuration ─────────────────────────────────────────────────────────
// Environment variables take precedence; values are read from the environment
// or .env.example (injected via Docker Compose / Kubernetes ConfigMap).

builder.Configuration.AddEnvironmentVariables();

// ── Redis ─────────────────────────────────────────────────────────────────
var redisConnection = builder.Configuration["REDIS_CONNECTION"] ?? "localhost:6379";

builder.Services.AddSingleton<IConnectionMultiplexer>(_ =>
{
    var opts = ConfigurationOptions.Parse(redisConnection);
    opts.AbortOnConnectFail = false;
    return ConnectionMultiplexer.Connect(opts);
});

// ── Application services ──────────────────────────────────────────────────
builder.Services.AddSingleton<IRedisService, RedisService>();
builder.Services.AddScoped<ICartService, CartServiceImpl>();

// ── ASP.NET Core ──────────────────────────────────────────────────────────
builder.Services.AddControllers();
builder.Services.AddOpenApi();

builder.Services.AddLogging(logging =>
{
    logging.AddConsole();
    logging.SetMinimumLevel(LogLevel.Information);
});

var app = builder.Build();

// ── Middleware ────────────────────────────────────────────────────────────
if (app.Environment.IsDevelopment())
    app.MapOpenApi();

app.UseRouting();
app.MapControllers();

app.Run();
