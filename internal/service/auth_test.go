package service

import (
	"context"
	"testing"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
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

func TestAuthService_Register(t *testing.T) {
	userRepo := &MockUserRepo{}
	passwordHasher := &MockPasswordHasher{}
	
	// Mock no existing user
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, assert.AnError)
	passwordHasher.On("Hash", mock.Anything, "password123").Return("hashed_password", nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	authService := &AuthService{
		userRepo:        userRepo,
		passwordManager: passwordHasher,
	}

	ctx := context.Background()
	user, err := authService.Register(ctx, "test@example.com", "password123", "Test User")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.FullName)
	assert.True(t, user.IsActive)

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
}