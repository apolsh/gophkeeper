package model

import (
	"errors"
	"fmt"
)

const (
	// Credentials secret item type for some credentials.
	Credentials string = "Credentials"
	// Text secret item type for some text.
	Text string = "Text"
	// Binary  secret item type for some file.
	Binary string = "Binary"
	// Card secret item type for some credit card.
	Card string = "Card"
)

// SecretItem item to be kept secure.
type SecretItem interface {
	// GetSecretPayload returns text implementation of secret item payload.
	GetSecretPayload() string
	// GetType returns secret item type.
	GetType() string
	// NewEncodedSecret encodes secret item.
	NewEncodedSecret(decodeFunction func(byteToDecode []byte) ([]byte, error), owner int64) (EncodedSecret, error)
}

// EncodedSecret encoded SecretItem.
type EncodedSecret struct {
	// ID SecretItem identifier.
	ID string
	// Name SecretItem name.
	Name string
	// Owner identifier SecretItem owner.
	Owner int64
	// Description of SecretItem.
	Description string
	// Type of SecretItem.
	Type string
	// EncodedContent of SecretItem.
	EncodedContent []byte
	// Hash of SecretItem.
	Hash string
	// Timestamp of last modification of SecretItem.
	Timestamp int64
}

// Decode decodes EncodedSecret.
func (e *EncodedSecret) Decode(decode func(byteToEncode []byte) ([]byte, error)) (SecretItem, error) {
	switch e.Type {
	case Credentials:
		return DecodeCredentialsSecretItem(decode, *e)
	case Text:
		return DecodeTextSecretItem(decode, *e)
	case Binary:
		return DecodeBinarySecretItem(decode, *e)
	case Card:
		return DecodeCardSecretItem(decode, *e)
	default:
		return nil, errors.New(fmt.Sprintf("failed to decode secret: unknown type %s", e.Type))
	}
}
