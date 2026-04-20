using Microsoft.EntityFrameworkCore;
using ReturnRefundService.Data;
using ReturnRefundService.Services;

var builder = WebApplication.CreateBuilder(args);

// ── Configuration ─────────────────────────────────────────────────────────
builder.Configuration.AddEnvironmentVariables();

// ── Database (PostgreSQL via EF Core) ─────────────────────────────────────
var databaseUrl = builder.Configuration["DATABASE_URL"]
    ?? "Host=localhost;Database=returns;Username=postgres;Password=postgres";

builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseNpgsql(databaseUrl, npgsqlOptions =>
    {
        npgsqlOptions.EnableRetryOnFailure(
            maxRetryCount: 5,
            maxRetryDelay: TimeSpan.FromSeconds(10),
            errorCodesToAdd: null);
    }));

// ── Application services ──────────────────────────────────────────────────
builder.Services.AddScoped<IReturnService, ReturnServiceImpl>();

// ── ASP.NET Core ──────────────────────────────────────────────────────────
builder.Services.AddControllers();
builder.Services.AddOpenApi();

builder.Services.AddLogging(logging =>
{
    logging.AddConsole();
    logging.SetMinimumLevel(LogLevel.Information);
});

var app = builder.Build();

// ── Auto-migrate on startup ───────────────────────────────────────────────
// In production you would run migrations as a separate init container.
// For local dev / preview environments, auto-migrate is convenient.
if (app.Environment.IsDevelopment())
{
    using var scope = app.Services.CreateScope();
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    await db.Database.MigrateAsync();
}

// ── Middleware ────────────────────────────────────────────────────────────
if (app.Environment.IsDevelopment())
    app.MapOpenApi();

app.UseRouting();
app.MapControllers();

app.Run();
