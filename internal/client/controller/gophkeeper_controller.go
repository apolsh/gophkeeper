package controller

import (
	"context"
	"errors"
	"fmt"
	"unicode"

	"github.com/apolsh/yapr-gophkeeper/internal/client/backend_client"
	"github.com/apolsh/yapr-gophkeeper/internal/client/storage"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
)

// IGophkeeperView UI for gophkeeper
type IGophkeeperView interface {
	// Show shows UI implementation
	Show(ctx context.Context) error
	// SetController sets controller for UI implementation
	SetController(controller *GophkeeperController)
	// SetAuthorized sets that user is already authorized for UI implementation
	SetAuthorized(isAuthorized bool)
	// ViewSecretsInfoList shows secret info list
	ViewSecretsInfoList(secretInfos []dto.SecretItemInfo)
	// ShowSecretItem shows secret item
	ShowSecretItem(item model.SecretItem)
	// GetStringInput gets input
	GetStringInput(ctx context.Context, inputText string) string
	// ShowError shows error
	ShowError(err error)
}

// LocalStorage local storage for gophkeeper
type LocalStorage interface {
	// SaveUser saves new user
	SaveUser(ctx context.Context, user model.User) error
	// UpdateUser update users metadata
	UpdateUser(ctx context.Context, user model.User) error
	// GetUserByID returns user by ID
	GetUserByID(ctx context.Context, userID int64) (model.User, error)
	// GetSecretSyncMetaByID returns metadata for one secret synchronization
	GetSecretSyncMetaByID(ctx context.Context, id string) (dto.SecretSyncMetadata, error)
	// GetSecretSyncMetaByOwnerID returns metadata for all secrets synchronization by user
	GetSecretSyncMetaByOwnerID(ctx context.Context, ownerID int64) ([]dto.SecretSyncMetadata, error)
	// SaveEncodedSecret saves EncodedSecret
	SaveEncodedSecret(ctx context.Context, encSecret model.EncodedSecret) error
	// GetSecretByID returns EncodedSecret by ID
	GetSecretByID(ctx context.Context, id string) (model.EncodedSecret, error)
	// GetAllSecretsItemInfoByUserID returns all secret item info by user id
	GetAllSecretsItemInfoByUserID(ctx context.Context, ownerID int64) ([]dto.SecretItemInfo, error)
	// GetSecretByName returns secret by it name
	GetSecretByName(ctx context.Context, name string) (model.EncodedSecret, error)
	// DeleteSecretByName delete secret by it name
	DeleteSecretByName(ctx context.Context, name string) (string, error)
}

// Encoder for decode and encode bytes
// should set secret key via SetSecretKey before use
type Encoder interface {
	// Encode encodes bytes
	Encode(byteToEncode []byte) ([]byte, error)
	// Decode decodes bytes
	Decode(byteToDecode []byte) ([]byte, error)
	// SetSecretKey sets secret key for encoder
	SetSecretKey(secretKey string) error
}

// BackendClient  client for interactions with backend
type BackendClient interface {
	// Login login user
	Login(ctx context.Context, login, password string) (string, model.User, error)
	// Register registers user
	Register(ctx context.Context, login, password string) (string, model.User, error)
	// SetAuthTokenForRequests sets authorization token for this client (add it to every required auth request)
	SetAuthTokenForRequests(token string)
	// GetSecretSyncMeta returns metadata for synchronization metadata
	GetSecretSyncMeta(ctx context.Context) ([]dto.SecretSyncMetadata, error)
	// GetSecretByID returns EncodedSecret by ID
	GetSecretByID(ctx context.Context, id string) (model.EncodedSecret, error)
	// SaveEncodedSecret  saves EncodedSecret
	SaveEncodedSecret(ctx context.Context, encSecret model.EncodedSecret) error
	// DeleteSecret delete EncodedSecret by ID
	DeleteSecret(ctx context.Context, id string) error
}

type authorizationMeta struct {
	id       int64
	login    string
	password string
}

// GophkeeperController core control for client gophkeeper application
type GophkeeperController struct {
	view          IGophkeeperView
	remoteStorage BackendClient
	authMeta      authorizationMeta
	localStorage  LocalStorage
	encoder       Encoder
}

// NewGophkeeperController GophkeeperController constructor
func NewGophkeeperController(
	ctx context.Context,
	view IGophkeeperView,
	backendClient BackendClient,
	localStorage LocalStorage,
	encoder Encoder,
	syncPeriod int64,

) *GophkeeperController {

	c := GophkeeperController{
		view:          view,
		remoteStorage: backendClient,
		localStorage:  localStorage,
		encoder:       encoder,
	}

	view.SetController(&c)
	runPeriodically(ctx, c.synchronizeSecretItems, view.ShowError, syncPeriod)

	return &c
}

// Login logins user
func (c *GophkeeperController) Login(ctx context.Context, login, password string) {
	token, user, err := c.remoteStorage.Login(ctx, login, password)
	if err != nil {
		c.view.ShowError(fmt.Errorf("failed to login user: %w", err))
		return
	}
	err = c.synchronizeAuthMeta(ctx, user)
	c.remoteStorage.SetAuthTokenForRequests(token)
	if err != nil {
		c.view.ShowError(err)
		return
	}
	c.authMeta = authorizationMeta{login: login, password: password, id: user.ID}
	err = c.synchronizeSecretItems(ctx)
	if err != nil {
		c.view.ShowError(err)
		c.authMeta = authorizationMeta{}
		return
	}
	err = c.encoder.SetSecretKey(password)
	if err != nil {
		c.view.ShowError(err)
		c.authMeta = authorizationMeta{}
		return
	}
	c.view.SetAuthorized(true)
}

// Register registers user
func (c *GophkeeperController) Register(ctx context.Context, login, password, repeatedPassword string) {
	err := passwordValidation(password, repeatedPassword)
	if err != nil {
		c.view.ShowError(err)
		return
	}
	token, user, err := c.remoteStorage.Register(ctx, login, password)
	if err != nil {
		c.view.ShowError(fmt.Errorf("failed to register user: %w", err))
		return
	}
	err = c.localStorage.SaveUser(ctx, user)
	if err != nil {
		c.view.ShowError(fmt.Errorf("failed to localy save user: %w", err))
		return
	}
	c.remoteStorage.SetAuthTokenForRequests(token)
	c.authMeta = authorizationMeta{login: login, password: password, id: user.ID}
	c.view.SetAuthorized(true)
	err = c.encoder.SetSecretKey(password)
	if err != nil {
		c.view.ShowError(err)
		return
	}
}

// SaveSecret saves secret item
func (c *GophkeeperController) SaveSecret(ctx context.Context, item model.SecretItem) {
	encodedSecret, err := item.NewEncodedSecret(c.encoder.Encode, c.authMeta.id)
	if err != nil {
		c.view.ShowError(fmt.Errorf("failed to encode secret: %w", err))
		return
	}
	err = c.localStorage.SaveEncodedSecret(ctx, encodedSecret)
	if err != nil {
		c.view.ShowError(fmt.Errorf("failed to locally store secret: %w", err))
		return
	}
	err = c.remoteStorage.SaveEncodedSecret(ctx, encodedSecret)
	if err != nil {
		if !errors.Is(backend_client.ErrServerIsNotAvailable, err) {
			c.view.ShowError(fmt.Errorf("synchronization operation failed: %w", err))
		}
		c.view.ShowError(err)
	}
}

// ListSecret shows all stored secrets of user
func (c *GophkeeperController) ListSecret(ctx context.Context) {
	secretItemsInfo, err := c.localStorage.GetAllSecretsItemInfoByUserID(ctx, c.authMeta.id)
	if err != nil {
		c.view.ShowError(fmt.Errorf("failed to get secrets info: %w", err))
		return
	}
	c.view.ViewSecretsInfoList(secretItemsInfo)
}

// GetSecret get decoded secret item by name
func (c *GophkeeperController) GetSecret(ctx context.Context, name string) {
	encodedSecret, err := c.localStorage.GetSecretByName(ctx, name)
	if err != nil {
		if errors.Is(storage.ErrorItemNotFound, err) {
			c.view.ShowError(fmt.Errorf("secret with name \"%s\" not found", name))
			return
		}
		c.view.ShowError(fmt.Errorf("failed to find secret %s: %w", name, err))
		return
	}
	secretItem, err := encodedSecret.Decode(c.encoder.Decode)
	if err != nil {
		c.view.ShowError(fmt.Errorf("failed to decode secret: %w", err))
		return
	}
	if secretItem.GetType() == model.Binary {
		path := c.view.GetStringInput(ctx, "please enter the path where the decoded file will be saved:")
		binarySecret, ok := secretItem.(*model.BinarySecretItem)
		if !ok {
			c.view.ShowError(fmt.Errorf("internal error: failed to complete type assertion"))
			return
		}
		err := binarySecret.SetOutputPath(path)
		if err != nil {
			c.view.ShowError(err)
			return
		}
	}

	c.view.ShowSecretItem(secretItem)
}

// DeleteSecret deletes secret item
func (c *GophkeeperController) DeleteSecret(ctx context.Context, name string) {
	id, err := c.localStorage.DeleteSecretByName(ctx, name)
	if err != nil {
		if errors.Is(storage.ErrorItemNotFound, err) {
			c.view.ShowError(fmt.Errorf("secret with name \"%s\" not found", name))
			return
		}
		c.view.ShowError(fmt.Errorf("failed to delete secret: %w", err))
		return
	}
	err = c.remoteStorage.DeleteSecret(ctx, id)
	if err != nil {
		if !errors.Is(backend_client.ErrServerIsNotAvailable, err) {
			c.view.ShowError(fmt.Errorf("synchronization operation failed: %w", err))
		}
		c.view.ShowError(err)
	}
}

// UnAuthorize ends the current user session
func (c *GophkeeperController) UnAuthorize() {
	c.authMeta = authorizationMeta{}
	c.view.SetAuthorized(false)
}

// Synchronize synchronize all secret metadata between client and backend
func (c *GophkeeperController) Synchronize(ctx context.Context) {
	err := c.synchronizeSecretItems(ctx)
	if err != nil {
		c.view.ShowError(err)
	}
}

func passwordValidation(password, repeatedPassword string) error {
	if password != repeatedPassword {
		return errors.New("passwords is not equal")
	}
	var size, number bool

	letters := 0
	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsLetter(c):
			letters++
		default:
		}
	}
	size = letters >= 3

	if !size || !number {
		return errors.New("password must contains at least 8 characters and 1 number")
	}

	return nil
}

func (c *GophkeeperController) synchronizeAuthMeta(ctx context.Context, user model.User) error {
	localUser, err := c.localStorage.GetUserByID(ctx, user.ID)
	if err != nil {
		if errors.Is(storage.ErrorItemNotFound, err) {
			return c.localStorage.SaveUser(ctx, user)
		}
		return err
	}
	if !user.EqualTo(localUser) {
		return c.localStorage.UpdateUser(ctx, user)
	}

	return nil
}

func (c *GophkeeperController) synchronizeSecretItems(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if c.authMeta.id != 0 {
			remoteSyncMetadata, err := c.remoteStorage.GetSecretSyncMeta(ctx)
			if err != nil {
				return err
			}
			localSyncMetadata, err := c.localStorage.GetSecretSyncMetaByOwnerID(ctx, c.authMeta.id)
			if err != nil {
				return err
			}
			localSyncMetadataMap := make(map[string]dto.SecretSyncMetadata)
			localIsSyncMap := make(map[string]bool)
			for _, meta := range localSyncMetadata {
				localSyncMetadataMap[meta.ID] = meta
				localIsSyncMap[meta.ID] = false
			}

			for _, remoteMeta := range remoteSyncMetadata {
				localMeta, contains := localSyncMetadataMap[remoteMeta.ID]
				if contains {
					localIsSyncMap[localMeta.ID] = true
					if remoteMeta.Hash != localMeta.Hash {
						if remoteMeta.Timestamp > localMeta.Timestamp {
							encodedSecretItem, err := c.remoteStorage.GetSecretByID(ctx, remoteMeta.ID)
							if err != nil {
								return err
							}
							err = c.localStorage.SaveEncodedSecret(ctx, encodedSecretItem)
							if err != nil {
								return err
							}
						} else {
							encodedSecretItem, err := c.localStorage.GetSecretByID(ctx, remoteMeta.ID)
							if err != nil {
								return err
							}
							err = c.remoteStorage.SaveEncodedSecret(ctx, encodedSecretItem)
							if err != nil {
								return err
							}
						}
					}
				} else {
					encodedSecretItem, err := c.remoteStorage.GetSecretByID(ctx, remoteMeta.ID)
					if err != nil {
						return err
					}
					err = c.localStorage.SaveEncodedSecret(ctx, encodedSecretItem)
					if err != nil {
						return err
					}
				}
			}

			for id, isSynchronized := range localIsSyncMap {
				if !isSynchronized {
					encodedSecretItem, err := c.localStorage.GetSecretByID(ctx, id)
					if err != nil {
						return err
					}
					err = c.remoteStorage.SaveEncodedSecret(ctx, encodedSecretItem)
					if err != nil {
						return err
					}
				}
			}
			return nil
		}
		return nil
	}
}
