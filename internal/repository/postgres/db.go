package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type DB struct {
	pool  *pgxpool.Pool
	sqlDB *sql.DB
}

func New(databaseURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Cloud SQL best practices
	cfg.MaxConns = 10
	cfg.MinConns = 1
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

	if lastErr != nil && pool == nil {
		return nil, errors.New("failed to connect to PostgreSQL after retries: " + lastErr.Error())
	}

	// Create sql.DB for explicit tx control
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Info().
		Int32("max_conns", cfg.MaxConns).
		Int32("min_conns", cfg.MinConns).
		Msg("PostgreSQL connection pool initialized")

	return &DB{
		pool:  pool,
		sqlDB: sqlDB,
	}, nil
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.sqlDB.BeginTx(ctx, nil)
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
	if db.sqlDB != nil {
		_ = db.sqlDB.Close()
	}
	log.Info().Msg("PostgreSQL connections closed")
}
