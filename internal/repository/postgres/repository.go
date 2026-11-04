package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peresvet668/Golang_Pr1/internal/domain"
)

//go:embed schema.sql
var schema embed.FS

type Repository struct {
	db *sql.DB
}

func Open(ctx context.Context, databaseURL string, maxOpenConnections int) (*Repository, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL connection: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxOpenConnections)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) Migrate(ctx context.Context) error {
	contents, err := schema.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read embedded schema: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, string(contents)); err != nil {
		return fmt.Errorf("apply database schema: %w", err)
	}
	return nil
}

func (r *Repository) EnsureDefaultWallet(ctx context.Context) error {
	const query = `
		INSERT INTO wallets (id, balance)
		VALUES ($1::uuid, 0)
		ON CONFLICT (id) DO NOTHING
	`
	if _, err := r.db.ExecContext(ctx, query, domain.DefaultWalletID); err != nil {
		return fmt.Errorf("ensure default wallet: %w", err)
	}
	return nil
}
