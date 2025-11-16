package service

import (
	"context"
	"testing"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Verify(ctx context.Context, password, hash string) (bool, error) {
	args := m.Called(ctx, password, hash)
	return args.Bool(0), args.Error(1)
}

type MockOrgRepo struct {
	mock.Mock
}

func (m *MockOrgRepo) Create(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	if args.Error(0) == nil {
		org.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrgRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Organization, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.Organization), args.Error(1)
}

func (m *MockOrgRepo) Update(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

type MockMembershipRepo struct {
	mock.Mock
}

func (m *MockMembershipRepo) Create(ctx context.Context, membership *model.OrgMembership) error {
	args := m.Called(ctx, membership)
	return args.Error(0)
}

func (m *MockMembershipRepo) GetByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*model.OrgMembership, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrgMembership), args.Error(1)
}

func (m *MockMembershipRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.OrgMembership, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.OrgMembership), args.Error(1)
}

func (m *MockMembershipRepo) Update(ctx context.Context, membership *model.OrgMembership) error {
	args := m.Called(ctx, membership)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	userRepo := &MockUserRepo{}
	orgRepo := &MockOrgRepo{}
	membershipRepo := &MockMembershipRepo{}
	passwordHasher := &MockPasswordHasher{}

	// Mock no existing user
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, pgx.ErrNoRows)
	passwordHasher.On("Hash", mock.Anything, "password123").Return("hashed_password", nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
	orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Organization")).Return(nil)
	membershipRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.OrgMembership")).Return(nil)

	authService := &AuthService{
		userRepo:        userRepo,
		orgRepo:         orgRepo,
		membershipRepo:  membershipRepo,
		passwordManager: passwordHasher,
	}

	ctx := context.Background()
	user, err := authService.Register(ctx, "test@example.com", "password123", "Test User")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.FullName)
	assert.True(t, user.IsActive)

	userRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
	membershipRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
}
