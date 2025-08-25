package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"marketstream/internal/config"
)

func ConnectDB(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	const (
		maxRetries = 10
		retryDelay = 2 * time.Second
		pingTTL    = 5 * time.Second
	)

	var db *sql.DB
	var err error

	for i := 1; i <= maxRetries; i++ {
		// открываем соединение
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			// ждём и пробуем снова
			select {
			case <-time.After(retryDelay):
				continue
			case <-ctx.Done():
				return nil, fmt.Errorf("db open canceled: %w", ctx.Err())
			}
		}

		// пингуем с таймаутом
		pctx, cancel := context.WithTimeout(ctx, pingTTL)
		err = db.PingContext(pctx)
		cancel()
		if err == nil {
			return db, nil // успех
		}

		_ = db.Close() // закрываем неудачное соединение

		// ретрай или отмена по контексту
		select {
		case <-time.After(retryDelay):
			continue
		case <-ctx.Done():
			return nil, fmt.Errorf("db ping canceled: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("database unreachable after %d attempts: %w", maxRetries, err)
}
