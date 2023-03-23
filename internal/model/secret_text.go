package model

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var _ SecretItem = (*TextSecretItem)(nil)

// TextSecretItem binary implementation of SecretItem
type TextSecretItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SecretType  string `json:"secretType"`
	Text        string `json:"text"`
}

// GetSecretPayload returns text implementation of secret item payload
func (c *TextSecretItem) GetSecretPayload() string {
	return fmt.Sprintf("[TEXT]: %s \n", c.Text)
}

// NewEncodedSecret encodes secret item
func (c *TextSecretItem) NewEncodedSecret(encodeFunction func(byteToDecode []byte) ([]byte, error), ownerID int64) (EncodedSecret, error) {
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
		Type:           Text,
		EncodedContent: encoded,
		Hash:           base64.StdEncoding.EncodeToString(hash[:]),
		Timestamp:      time.Now().UTC().UnixMilli(),
	}
	return encodedSecret, nil
}

// DecodeTextSecretItem decodes EncodedSecret item into TextSecretItem
func DecodeTextSecretItem(decode func(byteToEncode []byte) ([]byte, error), encoded EncodedSecret) (*TextSecretItem, error) {
	var textSecret TextSecretItem
	decodedBytes, err := decode(encoded.EncodedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode text secret: %w", err)
	}
	err = json.Unmarshal(decodedBytes, &textSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text secret: %w", err)
	}
	return &textSecret, nil
}

// NewTextSecretItem TextSecretItem constructor
func NewTextSecretItem(name, description, text string) *TextSecretItem {
	return &TextSecretItem{
		Name:        name,
		Description: description,
		Text:        text,
		SecretType:  Text,
	}
}

// GetType returns secret item type
func (c *TextSecretItem) GetType() string {
	return c.SecretType
}
