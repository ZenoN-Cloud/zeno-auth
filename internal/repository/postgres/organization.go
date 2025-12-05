package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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

func (r *OrganizationRepo) CreateTx(ctx context.Context, tx pgx.Tx, org *model.Organization) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO organizations (name, owner_user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	return tx.QueryRow(ctx, query, org.Name, org.OwnerUserID, org.Status, org.CreatedAt, org.UpdatedAt).Scan(&org.ID)
}

func (r *OrganizationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, name, owner_user_id, status, trial_ends_at, subscription_id, created_at, updated_at FROM organizations WHERE id = $1`

	org := &model.Organization{}
	err := r.db.pool.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.OwnerUserID, &org.Status, &org.TrialEndsAt, &org.SubscriptionID, &org.CreatedAt, &org.UpdatedAt)
	return org, err
}

func (r *OrganizationRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// First try: get organizations through membership
	query := `
		SELECT o.id, o.name, o.owner_user_id, o.status, o.country, o.currency, o.trial_ends_at, o.subscription_id, o.created_at, o.updated_at 
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
		if err := rows.Scan(&org.ID, &org.Name, &org.OwnerUserID, &org.Status, &org.Country, &org.Currency, &org.TrialEndsAt, &org.SubscriptionID, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// If no organizations found through membership, try direct ownership
	if len(orgs) == 0 {
		ownerQuery := `
			SELECT id, name, owner_user_id, status, country, currency, trial_ends_at, subscription_id, created_at, updated_at 
			FROM organizations 
			WHERE owner_user_id = $1`

		ownerRows, err := r.db.pool.Query(ctx, ownerQuery, userID)
		if err != nil {
			return nil, err
		}
		defer ownerRows.Close()

		for ownerRows.Next() {
			org := &model.Organization{}
			if err := ownerRows.Scan(&org.ID, &org.Name, &org.OwnerUserID, &org.Status, &org.Country, &org.Currency, &org.TrialEndsAt, &org.SubscriptionID, &org.CreatedAt, &org.UpdatedAt); err != nil {
				return nil, err
			}
			orgs = append(orgs, org)
		}

		if err := ownerRows.Err(); err != nil {
			return nil, err
		}
	}

	// Return empty slice if no organizations found (not an error)
	return orgs, nil
}

func (r *OrganizationRepo) CreateWithMembership(ctx context.Context, org *model.Organization, membership *model.OrgMembership) error {
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	orgQuery := `INSERT INTO organizations (name, owner_user_id, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	if err = tx.QueryRow(ctx, orgQuery, org.Name, org.OwnerUserID, org.Status, org.CreatedAt, org.UpdatedAt).Scan(&org.ID); err != nil {
		return err
	}

	membership.OrgID = org.ID
	membership.CreatedAt = now

	membershipQuery := `INSERT INTO org_memberships (user_id, org_id, role, is_active, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	if err = tx.QueryRow(ctx, membershipQuery, membership.UserID, membership.OrgID, membership.Role, membership.IsActive, membership.CreatedAt).Scan(&membership.ID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *OrganizationRepo) Update(ctx context.Context, org *model.Organization) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE organizations SET name = $2, status = $3, trial_ends_at = $4, subscription_id = $5, updated_at = $6 WHERE id = $1`

	org.UpdatedAt = time.Now()
	_, err := r.db.pool.Exec(ctx, query, org.ID, org.Name, org.Status, org.TrialEndsAt, org.SubscriptionID, org.UpdatedAt)
	return err
}
