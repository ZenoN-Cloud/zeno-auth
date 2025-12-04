package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(databaseURL string) (*DB, error) {
	// Validate database URL for path traversal
	if strings.Contains(databaseURL, "../") || strings.Contains(databaseURL, "..\\") {
		return nil, errors.New("invalid database URL: path traversal detected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Standard connection pool configuration
	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.HealthCheckPeriod = 30 * time.Second
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	var lastErr error

	// Retry strategy for Cloud SQL cold starts
	for i := 1; i <= 5; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil && pool != nil {
			err = pool.Ping(ctx)
		}

		if err == nil {
			log.Info().
				Int("attempt", i).
				Msg("Connected to PostgreSQL via pgxpool")
			break
		}

		lastErr = err
		log.Warn().
			Err(err).
			Int("attempt", i).
			Msg("PostgreSQL connection failed, retrying...")

		time.Sleep(time.Duration(i) * time.Second)
	}

	if pool == nil {
		return nil, errors.New("failed to connect to PostgreSQL after retries: " + lastErr.Error())
	}

	log.Info().
		Int32("max_conns", cfg.MaxConns).
		Int32("min_conns", cfg.MinConns).
		Msg("PostgreSQL connection pool initialized")

	return &DB{
		pool: pool,
	}, nil
}

func (db *DB) Pool() *pgxpool.Pool {
	if db == nil {
		return nil
	}
	return db.pool
}

func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if db == nil || db.pool == nil {
		return nil, errors.New("database connection is nil")
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return nil, err
	}

	return tx, nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
		log.Info().Msg("PostgreSQL connection closed")
	}
}
