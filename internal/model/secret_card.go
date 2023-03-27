package model

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var _ SecretItem = (*CardSecretItem)(nil)

// CardSecretItem binary implementation of SecretItem.
type CardSecretItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SecretType  string `json:"secretType"`
	OwnerName   string `json:"owner"`
	Number      string `json:"number"`
	CVV         string `json:"cvv"`
}

// GetType returns secret item type.
func (c *CardSecretItem) GetType() string {
	return c.SecretType
}

// GetSecretPayload returns text implementation of secret item payload.
func (c *CardSecretItem) GetSecretPayload() string {
	return fmt.Sprintf("[OWNER]: %s \n[NUMBER]: %s \n[CVV]: %s \n", c.OwnerName, c.Number, c.CVV)
}

// NewEncodedSecret encodes secret item.
func (c *CardSecretItem) NewEncodedSecret(encodeFunction func(byteToDecode []byte) ([]byte, error), ownerID int64) (EncodedSecret, error) {
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
		Type:           Card,
		EncodedContent: encoded,
		Hash:           base64.StdEncoding.EncodeToString(hash[:]),
		Timestamp:      time.Now().UTC().UnixMilli(),
	}
	return encodedSecret, nil
}

// DecodeCardSecretItem decodes EncodedSecret item into CardSecretItem.
func DecodeCardSecretItem(decode func(byteToEncode []byte) ([]byte, error), encoded EncodedSecret) (*CardSecretItem, error) {
	var credentialsSecret CardSecretItem
	decodedBytes, err := decode(encoded.EncodedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode card secret: %w", err)
	}
	err = json.Unmarshal(decodedBytes, &credentialsSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse card secret: %w", err)
	}
	return &credentialsSecret, nil
}

// NewCardSecretItem CardSecretItem constructor.
func NewCardSecretItem(name, description, owner, number, cvv string) *CardSecretItem {
	return &CardSecretItem{
		Name:        name,
		Description: description,
		OwnerName:   owner,
		Number:      number,
		CVV:         cvv,
		SecretType:  Card,
	}
}
