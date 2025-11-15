package postgres

import (
	"context"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
)

type MembershipRepo struct {
	db *DB
}

func NewMembershipRepo(db *DB) *MembershipRepo {
	return &MembershipRepo{db: db}
}

func (r *MembershipRepo) Create(ctx context.Context, membership *model.OrgMembership) error {
	query := `
		INSERT INTO org_memberships (user_id, org_id, role, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	
	membership.CreatedAt = time.Now()
	
	return r.db.pool.QueryRow(ctx, query, membership.UserID, membership.OrgID, membership.Role, membership.IsActive, membership.CreatedAt).Scan(&membership.ID)
}

func (r *MembershipRepo) GetByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*model.OrgMembership, error) {
	query := `SELECT id, user_id, org_id, role, is_active, created_at FROM org_memberships WHERE user_id = $1 AND org_id = $2`
	
	membership := &model.OrgMembership{}
	err := r.db.pool.QueryRow(ctx, query, userID, orgID).Scan(&membership.ID, &membership.UserID, &membership.OrgID, &membership.Role, &membership.IsActive, &membership.CreatedAt)
	return membership, err
}

func (r *MembershipRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.OrgMembership, error) {
	query := `SELECT id, user_id, org_id, role, is_active, created_at FROM org_memberships WHERE user_id = $1 AND is_active = true`
	
	rows, err := r.db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*model.OrgMembership
	for rows.Next() {
		membership := &model.OrgMembership{}
		if err := rows.Scan(&membership.ID, &membership.UserID, &membership.OrgID, &membership.Role, &membership.IsActive, &membership.CreatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}

	return memberships, rows.Err()
}

func (r *MembershipRepo) Update(ctx context.Context, membership *model.OrgMembership) error {
	query := `UPDATE org_memberships SET role = $3, is_active = $4 WHERE user_id = $1 AND org_id = $2`
	
	_, err := r.db.pool.Exec(ctx, query, membership.UserID, membership.OrgID, membership.Role, membership.IsActive)
	return err
}