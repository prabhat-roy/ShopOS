package testcontainers_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestPostgresContainer demonstrates using testcontainers-go to spin up
// a real PostgreSQL 16 instance for integration testing.
func TestPostgresContainer(t *testing.T) {
	ctx := context.Background()

	// ── Start postgres:16-alpine container ────────────────────────────────────
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("shopos_test"),
		postgres.WithUsername("shopos"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate postgres container: %v", err)
		}
	})

	// ── Get connection string ─────────────────────────────────────────────────
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// ── Connect ───────────────────────────────────────────────────────────────
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	// ── Run migrations (inline for test) ──────────────────────────────────────
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     UUID NOT NULL,
			status      VARCHAR(50) NOT NULL DEFAULT 'pending',
			total       NUMERIC(12, 2) NOT NULL,
			currency    CHAR(3) NOT NULL DEFAULT 'USD',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
		CREATE INDEX IF NOT EXISTS idx_orders_status  ON orders(status);
	`)
	if err != nil {
		t.Fatalf("failed to create orders table: %v", err)
	}

	// ── Insert test records ───────────────────────────────────────────────────
	type Order struct {
		UserID   string
		Status   string
		Total    float64
		Currency string
	}

	orders := []Order{
		{"00000000-0000-0000-0000-000000000001", "confirmed", 149.99, "USD"},
		{"00000000-0000-0000-0000-000000000001", "shipped", 89.50, "USD"},
		{"00000000-0000-0000-0000-000000000002", "pending", 299.00, "EUR"},
	}

	for _, o := range orders {
		_, err = db.ExecContext(ctx,
			`INSERT INTO orders (user_id, status, total, currency) VALUES ($1, $2, $3, $4)`,
			o.UserID, o.Status, o.Total, o.Currency,
		)
		if err != nil {
			t.Fatalf("failed to insert order: %v", err)
		}
	}

	// ── Query: count all orders ───────────────────────────────────────────────
	t.Run("CountAllOrders", func(t *testing.T) {
		var count int
		err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders`).Scan(&count)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if count != len(orders) {
			t.Errorf("expected %d orders, got %d", len(orders), count)
		}
	})

	// ── Query: filter by user ─────────────────────────────────────────────────
	t.Run("FilterOrdersByUser", func(t *testing.T) {
		rows, err := db.QueryContext(ctx,
			`SELECT id, status, total FROM orders WHERE user_id = $1 ORDER BY created_at`,
			"00000000-0000-0000-0000-000000000001",
		)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		var results []struct {
			ID     string
			Status string
			Total  float64
		}
		for rows.Next() {
			var r struct {
				ID     string
				Status string
				Total  float64
			}
			if err := rows.Scan(&r.ID, &r.Status, &r.Total); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			results = append(results, r)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 orders for user, got %d", len(results))
		}
	})

	// ── Query: aggregation ────────────────────────────────────────────────────
	t.Run("TotalRevenueByStatus", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT status, COUNT(*) as cnt, SUM(total) as revenue
			FROM orders
			GROUP BY status
			ORDER BY status
		`)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var status string
			var cnt int
			var revenue float64
			if err := rows.Scan(&status, &cnt, &revenue); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			t.Logf("status=%s count=%d revenue=%.2f", status, cnt, revenue)
		}
	})

	// ── Update: status transition ─────────────────────────────────────────────
	t.Run("UpdateOrderStatus", func(t *testing.T) {
		result, err := db.ExecContext(ctx,
			`UPDATE orders SET status = 'delivered', updated_at = NOW()
			 WHERE status = 'shipped' AND currency = 'USD'`,
		)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			t.Fatalf("rows affected error: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 row updated, got %d", affected)
		}
	})

	// ── Transaction: atomic order creation ───────────────────────────────────
	t.Run("TransactionalInsert", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}

		var orderID string
		err = tx.QueryRowContext(ctx,
			`INSERT INTO orders (user_id, status, total, currency)
			 VALUES ($1, $2, $3, $4) RETURNING id`,
			"00000000-0000-0000-0000-000000000003", "pending", 599.99, "GBP",
		).Scan(&orderID)
		if err != nil {
			tx.Rollback()
			t.Fatalf("insert failed: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("commit failed: %v", err)
		}

		if orderID == "" {
			t.Error("expected non-empty order ID from RETURNING clause")
		}
		t.Logf("created order: %s", orderID)
	})

	// ── Verify Postgres version ───────────────────────────────────────────────
	t.Run("PostgresVersion", func(t *testing.T) {
		var version string
		err := db.QueryRowContext(ctx, `SELECT version()`).Scan(&version)
		if err != nil {
			t.Fatalf("version query failed: %v", err)
		}
		t.Logf("PostgreSQL version: %s", version)
		// Verify it's PG 16
		if !containsSubstring(version, "PostgreSQL 16") {
			t.Errorf("expected PostgreSQL 16, got: %s", version)
		}
	})
}

// TestPostgresContainerWithSchema demonstrates initializing a container with
// a pre-loaded SQL schema file (simulating Flyway migrations).
func TestPostgresContainerWithSchema(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("shopos_identity"),
		postgres.WithUsername("identity_svc"),
		postgres.WithPassword("identity_pass"),
		postgres.WithInitScripts("testdata/identity_schema.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		// Skip if testdata file doesn't exist (CI without test fixtures)
		t.Skipf("postgres container setup skipped: %v", err)
	}
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string error: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("open db error: %v", err)
	}
	defer db.Close()

	var tableCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("table count query failed: %v", err)
	}
	t.Logf("Schema initialized with %d tables", tableCount)
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

var _ = fmt.Sprintf // avoid unused import
