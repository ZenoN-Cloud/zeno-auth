package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type DB struct {
	pool  *pgxpool.Pool
	sqlDB *sql.DB
}

func New(databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	// Also create sql.DB for transactions
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Info().Msg("Connected to PostgreSQL")
	return &DB{pool: pool, sqlDB: sqlDB}, nil
}

func (db *DB) Close() {
	db.pool.Close()
	if db.sqlDB != nil {
		_ = db.sqlDB.Close()
	}
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.sqlDB.BeginTx(ctx, nil)
}
