//go:generate mockgen -destination=../../mocks/backend_gophkeeper_service_mocks.go -package=mocks github.com/apolsh/yapr-gophkeeper/internal/backend/service UserStorage,SecretStorage
package service

import (
	"context"
	"errors"
	"fmt"

	tokenManager "github.com/apolsh/yapr-gophkeeper/internal/backend/token_manager"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	errs "github.com/apolsh/yapr-gophkeeper/internal/model/app_errors"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
	"golang.org/x/crypto/bcrypt"
)

// UserStorage storage of users.
type UserStorage interface {
	// NewUser creates new user
	NewUser(ctx context.Context, login string, s string) (model.User, error)
	// GetUserByLogin returns user by login
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	// Close for graceful shutdown
	Close()
}

// SecretStorage storage for EncodedSecret.
type SecretStorage interface {
	// GetSecretSyncMetaByUser returns metadata for secret synchronization
	GetSecretSyncMetaByUser(ctx context.Context, userID int64) ([]dto.SecretSyncMetadata, error)
	// GetSecretSyncMetaByOwnerAndName returns metadata for secret synchronization by ownerID and secret name
	GetSecretSyncMetaByOwnerAndName(ctx context.Context, userID int, name string) (dto.SecretSyncMetadata, error)
	// GetSecretByID returns EncodedSecret by ID
	GetSecretByID(ctx context.Context, userID int, secretID string) (model.EncodedSecret, error)
	// SaveEncodedSecret saves EncodedSecret
	SaveEncodedSecret(ctx context.Context, secret model.EncodedSecret) error
	// DeleteEncodedSecret deletes EncodedSecret
	DeleteEncodedSecret(ctx context.Context, ownerID int, secretID string) error
	// Close for graceful shutdown
	Close()
}

var (
	// ErrUserNotFound user not found error.
	ErrUserNotFound = errors.New("the specified user is not registered in the system")
	// ErrOwnerMissmatch owner missmatch error.
	ErrOwnerMissmatch = errors.New("requested item belongs to another user")
)

// GophkeeperServiceImpl service for EncodedSecret and User management.
type GophkeeperServiceImpl struct {
	tokenManager  tokenManager.TokenManager
	userStorage   UserStorage
	secretStorage SecretStorage
}

// NewGophkeeperService GophkeeperServiceImpl constructor.
func NewGophkeeperService(tokenManager tokenManager.TokenManager, userStorage UserStorage, secretStorage SecretStorage) *GophkeeperServiceImpl {
	return &GophkeeperServiceImpl{
		tokenManager:  tokenManager,
		userStorage:   userStorage,
		secretStorage: secretStorage,
	}
}

// Login login user.
func (s *GophkeeperServiceImpl) Login(ctx context.Context, login string, password string) (string, model.User, error) {
	if login == "" || password == "" {
		return "", model.User{}, errs.ErrorEmptyValue
	}
	user, err := s.userStorage.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(errs.ErrItemNotFound, err) {
			return "", model.User{}, ErrUserNotFound
		}
		return "", model.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return "", model.User{}, errs.ErrorInvalidPassword
	}

	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return "", model.User{}, fmt.Errorf("failed to generate token : %w", err)
	}

	return token, user, nil
}

// Register register user.
func (s *GophkeeperServiceImpl) Register(ctx context.Context, login string, password string) (string, model.User, error) {
	if login == "" || password == "" {
		return "", model.User{}, errs.ErrorEmptyValue
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", model.User{}, err
	}

	user, err := s.userStorage.NewUser(ctx, login, string(hashedPassword))
	if err != nil {
		return "", model.User{}, err
	}
	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return "", model.User{}, fmt.Errorf("failed to generate token : %w", err)
	}

	return token, user, nil
}

// GetSecretSyncMetaByUser returns metadata for secret synchronization.
func (s *GophkeeperServiceImpl) GetSecretSyncMetaByUser(ctx context.Context, id int64) ([]dto.SecretSyncMetadata, error) {
	return s.secretStorage.GetSecretSyncMetaByUser(ctx, id)
}

// GetSecretSyncMetaByOwnerAndName returns metadata for secret synchronization by owner and name.
func (s *GophkeeperServiceImpl) GetSecretSyncMetaByOwnerAndName(ctx context.Context, userID int, name string) (dto.SecretSyncMetadata, error) {
	return s.secretStorage.GetSecretSyncMetaByOwnerAndName(ctx, userID, name)
}

// GetSecret returns EncodedSecret by ID
func (s *GophkeeperServiceImpl) GetSecret(ctx context.Context, userID int, secretID string) (model.EncodedSecret, error) {
	encodedSecret, err := s.secretStorage.GetSecretByID(ctx, userID, secretID)
	if err != nil {
		return model.EncodedSecret{}, err
	}
	return encodedSecret, nil
}

// SaveEncodedSecret saves EncodedSecret.
func (s *GophkeeperServiceImpl) SaveEncodedSecret(ctx context.Context, ownerID int, secret model.EncodedSecret) error {
	if int64(ownerID) != secret.Owner {
		return ErrOwnerMissmatch
	}

	return s.secretStorage.SaveEncodedSecret(ctx, secret)
}

// DeleteSecret delete EncodedSecret by ID.
func (s *GophkeeperServiceImpl) DeleteSecret(ctx context.Context, ownerID int, secretID string) error {
	return s.secretStorage.DeleteEncodedSecret(ctx, ownerID, secretID)
}
