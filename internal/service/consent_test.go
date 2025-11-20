package service

import (
	"context"
	"testing"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockConsentRepository struct {
	mock.Mock
}

func (m *MockConsentRepository) Create(ctx context.Context, consent *model.UserConsent) error {
	args := m.Called(ctx, consent)
	return args.Error(0)
}

func (m *MockConsentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.UserConsent, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserConsent), args.Error(1)
}

func (m *MockConsentRepository) GetByUserAndType(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) (*model.UserConsent, error) {
	args := m.Called(ctx, userID, consentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserConsent), args.Error(1)
}

func (m *MockConsentRepository) Revoke(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) error {
	args := m.Called(ctx, userID, consentType)
	return args.Error(0)
}

func TestConsentService_GrantConsent(t *testing.T) {
	mockRepo := new(MockConsentRepository)
	service := NewConsentService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserAndType", ctx, userID, model.ConsentTypeTerms).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.UserConsent")).Return(nil)

	err := service.GrantConsent(ctx, userID, model.ConsentTypeTerms, "1.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestConsentService_RevokeConsent(t *testing.T) {
	mockRepo := new(MockConsentRepository)
	service := NewConsentService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("Revoke", ctx, userID, model.ConsentTypeMarketing).Return(nil)

	err := service.RevokeConsent(ctx, userID, model.ConsentTypeMarketing)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestConsentService_HasConsent(t *testing.T) {
	mockRepo := new(MockConsentRepository)
	service := NewConsentService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	consent := &model.UserConsent{
		ID:          uuid.New(),
		UserID:      userID,
		ConsentType: model.ConsentTypePrivacy,
		Version:     "1.0",
		Granted:     true,
		GrantedAt:   time.Now(),
	}

	mockRepo.On("GetByUserAndType", ctx, userID, model.ConsentTypePrivacy).Return(consent, nil)

	hasConsent, err := service.HasConsent(ctx, userID, model.ConsentTypePrivacy)
	assert.NoError(t, err)
	assert.True(t, hasConsent)
	mockRepo.AssertExpectations(t)
}
