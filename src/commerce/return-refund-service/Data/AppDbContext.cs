using Microsoft.EntityFrameworkCore;
using ReturnRefundService.Models;

namespace ReturnRefundService.Data;

public class AppDbContext : DbContext
{
    public AppDbContext(DbContextOptions<AppDbContext> options) : base(options) { }

    public DbSet<ReturnRequest> ReturnRequests => Set<ReturnRequest>();
    public DbSet<RefundRecord> RefundRecords => Set<RefundRecord>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        // ── ReturnRequest ────────────────────────────────────────────────
        modelBuilder.Entity<ReturnRequest>(entity =>
        {
            entity.ToTable("return_requests");
            entity.HasKey(e => e.Id);

            entity.Property(e => e.Id)
                  .HasColumnName("id")
                  .HasDefaultValueSql("gen_random_uuid()");

            entity.Property(e => e.OrderId)
                  .HasColumnName("order_id")
                  .IsRequired()
                  .HasMaxLength(100);

            entity.Property(e => e.CustomerId)
                  .HasColumnName("customer_id")
                  .IsRequired()
                  .HasMaxLength(100);

            entity.Property(e => e.ProductId)
                  .HasColumnName("product_id")
                  .IsRequired()
                  .HasMaxLength(100);

            entity.Property(e => e.Quantity)
                  .HasColumnName("quantity");

            entity.Property(e => e.Reason)
                  .HasColumnName("reason")
                  .HasConversion<string>();

            entity.Property(e => e.Notes)
                  .HasColumnName("notes")
                  .HasMaxLength(2000);

            entity.Property(e => e.Status)
                  .HasColumnName("status")
                  .HasConversion<string>();

            entity.Property(e => e.CreatedAt)
                  .HasColumnName("created_at")
                  .HasDefaultValueSql("NOW()");

            entity.Property(e => e.UpdatedAt)
                  .HasColumnName("updated_at")
                  .HasDefaultValueSql("NOW()");

            entity.HasIndex(e => e.CustomerId)
                  .HasDatabaseName("idx_return_requests_customer_id");

            entity.HasIndex(e => e.OrderId)
                  .HasDatabaseName("idx_return_requests_order_id");

            entity.HasIndex(e => e.Status)
                  .HasDatabaseName("idx_return_requests_status");
        });

        // ── RefundRecord ─────────────────────────────────────────────────
        modelBuilder.Entity<RefundRecord>(entity =>
        {
            entity.ToTable("refund_records");
            entity.HasKey(e => e.Id);

            entity.Property(e => e.Id)
                  .HasColumnName("id")
                  .HasDefaultValueSql("gen_random_uuid()");

            entity.Property(e => e.ReturnRequestId)
                  .HasColumnName("return_request_id")
                  .IsRequired();

            entity.Property(e => e.Amount)
                  .HasColumnName("amount")
                  .HasPrecision(12, 2);

            entity.Property(e => e.Currency)
                  .HasColumnName("currency")
                  .HasMaxLength(3)
                  .HasDefaultValue("USD");

            entity.Property(e => e.Method)
                  .HasColumnName("method")
                  .HasMaxLength(50)
                  .HasDefaultValue("original");

            entity.Property(e => e.ProcessedAt)
                  .HasColumnName("processed_at")
                  .HasDefaultValueSql("NOW()");

            entity.HasOne(e => e.ReturnRequest)
                  .WithOne(r => r.Refund)
                  .HasForeignKey<RefundRecord>(e => e.ReturnRequestId)
                  .OnDelete(DeleteBehavior.Cascade);

            entity.HasIndex(e => e.ReturnRequestId)
                  .IsUnique()
                  .HasDatabaseName("idx_refund_records_return_request_id");
        });
    }
}
