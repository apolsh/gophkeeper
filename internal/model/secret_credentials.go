package model

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var _ SecretItem = (*CredentialsSecretItem)(nil)

// CredentialsSecretItem binary implementation of SecretItem.
type CredentialsSecretItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SecretType  string `json:"secretType"`
	Login       string `json:"login"`
	Password    string `json:"password"`
}

// GetType returns secret item type.
func (c *CredentialsSecretItem) GetType() string {
	return c.SecretType
}

// GetSecretPayload returns text implementation of secret item payload.
func (c *CredentialsSecretItem) GetSecretPayload() string {
	return fmt.Sprintf("[LOGIN]: %s \n[PASSWORD]: %s \n", c.Login, c.Password)
}

// NewEncodedSecret encodes secret item.
func (c *CredentialsSecretItem) NewEncodedSecret(encodeFunction func(byteToDecode []byte) ([]byte, error), ownerID int64) (EncodedSecret, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return EncodedSecret{}, err
	}
	encoded, err := encodeFunction(jsonBytes)
	if err != nil {
		return EncodedSecret{}, err
	}
	hash := sha256.Sum256(jsonBytes)

	encodedSecret := EncodedSecret{
		ID:             uuid.New().String(),
		Name:           c.Name,
		Owner:          ownerID,
		Description:    c.Description,
		Type:           Credentials,
		EncodedContent: encoded,
		Hash:           base64.StdEncoding.EncodeToString(hash[:]),
		Timestamp:      time.Now().UTC().UnixMilli(),
	}
	return encodedSecret, nil
}

// DecodeCredentialsSecretItem decodes EncodedSecret item into CredentialsSecretItem.
func DecodeCredentialsSecretItem(decode func(byteToEncode []byte) ([]byte, error), encoded EncodedSecret) (*CredentialsSecretItem, error) {
	var credentialsSecret CredentialsSecretItem
	decodedBytes, err := decode(encoded.EncodedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode credentials secret: %w", err)
	}
	err = json.Unmarshal(decodedBytes, &credentialsSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials secret: %w", err)
	}
	return &credentialsSecret, nil
}

// NewCredentialsSecretItem CredentialsSecretItem constructor.
func NewCredentialsSecretItem(name, description, login, password string) *CredentialsSecretItem {
	return &CredentialsSecretItem{
		Name:        name,
		Description: description,
		Login:       login,
		Password:    password,
		SecretType:  Credentials,
	}
}
