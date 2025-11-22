package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

type OrganizationRepo struct {
	db *DB
}

func NewOrganizationRepo(db *DB) *OrganizationRepo {
	return &OrganizationRepo{db: db}
}

func (r *OrganizationRepo) Create(ctx context.Context, org *model.Organization) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO organizations (name, owner_user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	return r.db.pool.QueryRow(ctx, query, org.Name, org.OwnerUserID, org.Status, org.CreatedAt, org.UpdatedAt).Scan(&org.ID)
}

func (r *OrganizationRepo) CreateTx(ctx context.Context, tx *sql.Tx, org *model.Organization) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO organizations (name, owner_user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	return tx.QueryRowContext(ctx, query, org.Name, org.OwnerUserID, org.Status, org.CreatedAt, org.UpdatedAt).Scan(&org.ID)
}

func (r *OrganizationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, name, owner_user_id, status, created_at, updated_at FROM organizations WHERE id = $1`

	org := &model.Organization{}
	err := r.db.pool.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.OwnerUserID, &org.Status, &org.CreatedAt, &org.UpdatedAt)
	return org, err
}

func (r *OrganizationRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT o.id, o.name, o.owner_user_id, o.status, o.created_at, o.updated_at 
		FROM organizations o
		JOIN org_memberships m ON o.id = m.org_id
		WHERE m.user_id = $1 AND m.is_active = true`

	rows, err := r.db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.OwnerUserID, &org.Status, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty slice if no organizations found (not an error)
	return orgs, nil
}

func (r *OrganizationRepo) Update(ctx context.Context, org *model.Organization) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE organizations SET name = $2, status = $3, updated_at = $4 WHERE id = $1`

	org.UpdatedAt = time.Now()
	_, err := r.db.pool.Exec(ctx, query, org.ID, org.Name, org.Status, org.UpdatedAt)
	return err
}
