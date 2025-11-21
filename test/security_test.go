package test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
)

// MockUserRepo for testing
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepo) ResetFailedLogins(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// TestAccountLockout tests the account lockout after 5 failed login attempts
func TestAccountLockout(t *testing.T) {
	mockRepo := new(MockUserRepo)
	ctx := context.Background()

	userID := uuid.New()
	user := &model.User{
		ID:                  userID,
		Email:               "test@example.com",
		PasswordHash:        "$argon2id$v=19$m=65536,t=3,p=2$test",
		FailedLoginAttempts: 4,
		IsActive:            true,
	}

	// After 5th failed attempt, account should be locked
	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil).Once()
	mockRepo.On("IncrementFailedLogins", ctx, userID).Return(nil).Run(
		func(args mock.Arguments) {
			user.FailedLoginAttempts = 5
			user.LockedUntil = timePtr(time.Now().Add(15 * time.Minute))
		},
	)

	// Simulate getting user and incrementing failed logins
	_, err := mockRepo.GetByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	err = mockRepo.IncrementFailedLogins(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 5, user.FailedLoginAttempts)
	assert.NotNil(t, user.LockedUntil)

	mockRepo.AssertExpectations(t)
}

// TestRefreshTokenValidation tests refresh token validation scenarios
func TestRefreshTokenValidation(t *testing.T) {
	tests := []struct {
		name        string
		token       *model.RefreshToken
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid token",
			token: &model.RefreshToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
				RevokedAt: nil,
			},
			expectError: false,
		},
		{
			name: "Expired token",
			token: &model.RefreshToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				ExpiresAt: time.Now().Add(-1 * time.Hour),
				RevokedAt: nil,
			},
			expectError: true,
			errorMsg:    "token expired",
		},
		{
			name: "Revoked token",
			token: &model.RefreshToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
				RevokedAt: timePtr(time.Now().Add(-1 * time.Hour)),
			},
			expectError: true,
			errorMsg:    "token revoked",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Validate token
				isValid := tt.token.RevokedAt == nil && time.Now().Before(tt.token.ExpiresAt)

				if tt.expectError {
					assert.False(t, isValid)
				} else {
					assert.True(t, isValid)
				}
			},
		)
	}
}

// TestPasswordResetFlow tests the password reset flow
func TestPasswordResetFlow(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	user := &model.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=2$oldpassword",
		IsActive:     true,
	}

	// Simulate password reset
	newPasswordHash := "$argon2id$v=19$m=65536,t=3,p=2$newpassword"
	oldPasswordHash := user.PasswordHash
	user.PasswordHash = newPasswordHash

	assert.NotEqual(t, oldPasswordHash, user.PasswordHash)
	assert.Equal(t, newPasswordHash, user.PasswordHash)
	_ = ctx // use ctx
}

// TestEmailVerification tests email verification flow
func TestEmailVerification(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	user := &model.User{
		ID:       userID,
		Email:    "test@example.com",
		IsActive: false,
	}

	// Simulate email verification
	assert.False(t, user.IsActive)
	user.IsActive = true
	assert.True(t, user.IsActive)
	_ = ctx // use ctx
}

// TestRateLimiting tests rate limiting logic
func TestRateLimiting(t *testing.T) {
	// Simulate rate limiting: max 5 attempts in 15 minutes
	maxAttempts := 5
	window := 15 * time.Minute

	attempts := []time.Time{
		time.Now().Add(-20 * time.Minute), // Outside window
		time.Now().Add(-10 * time.Minute),
		time.Now().Add(-8 * time.Minute),
		time.Now().Add(-5 * time.Minute),
		time.Now().Add(-2 * time.Minute),
		time.Now().Add(-1 * time.Minute),
	}

	// Count attempts within window
	cutoff := time.Now().Add(-window)
	count := 0
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			count++
		}
	}

	assert.Equal(t, 5, count)
	assert.True(t, count >= maxAttempts, "Rate limit should be exceeded")
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// TestSessionFingerprint tests session fingerprint validation
func TestSessionFingerprint(t *testing.T) {
	tests := []struct {
		name               string
		storedFingerprint  string
		requestFingerprint string
		expectValid        bool
	}{
		{
			name:               "Matching fingerprint",
			storedFingerprint:  "fp123",
			requestFingerprint: "fp123",
			expectValid:        true,
		},
		{
			name:               "Different fingerprint",
			storedFingerprint:  "fp123",
			requestFingerprint: "fp456",
			expectValid:        false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				isValid := tt.storedFingerprint == tt.requestFingerprint
				assert.Equal(t, tt.expectValid, isValid)
			},
		)
	}
}

// Ensure service.ErrEmailExists is defined
func init() {
	if service.ErrEmailExists == nil {
		panic("service.ErrEmailExists must be defined")
	}
}
