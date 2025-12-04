package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (email, password_hash, full_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return r.db.pool.QueryRow(
		ctx, query, user.Email, user.PasswordHash, user.FullName, user.IsActive, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *UserRepo) CreateTx(ctx context.Context, tx pgx.Tx, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (email, password_hash, full_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return tx.QueryRow(
		ctx, query, user.Email, user.PasswordHash, user.FullName, user.IsActive, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, email, password_hash, full_name, is_active, failed_login_attempts, locked_until, created_at, updated_at FROM users WHERE id = $1`

	user := &model.User{}
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.IsActive, &user.FailedLoginAttempts,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, email, password_hash, full_name, is_active, failed_login_attempts, locked_until, created_at, updated_at FROM users WHERE email = $1`

	user := &model.User{}
	err := r.db.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.IsActive, &user.FailedLoginAttempts,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE users SET email = $2, password_hash = $3, full_name = $4, is_active = $5, failed_login_attempts = $6, locked_until = $7, updated_at = $8 WHERE id = $1`

	user.UpdatedAt = time.Now()
	_, err := r.db.pool.Exec(
		ctx, query, user.ID, user.Email, user.PasswordHash, user.FullName, user.IsActive, user.FailedLoginAttempts,
		user.LockedUntil, user.UpdatedAt,
	)
	return err
}

func (r *UserRepo) UpdateTx(ctx context.Context, tx pgx.Tx, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE users SET email = $2, password_hash = $3, full_name = $4, is_active = $5, failed_login_attempts = $6, locked_until = $7, updated_at = $8 WHERE id = $1`

	user.UpdatedAt = time.Now()
	_, err := tx.Exec(
		ctx, query, user.ID, user.Email, user.PasswordHash, user.FullName, user.IsActive, user.FailedLoginAttempts,
		user.LockedUntil, user.UpdatedAt,
	)
	return err
}
