package encoder

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/apolsh/yapr-gophkeeper/internal/client/controller"
)

// AESGMCEncoder implementation of Encoder with AES with Galois/Counter Mode (AES-GCM).
type AESGMCEncoder struct {
	ready   bool
	encoder cipher.AEAD
	nonce   []byte
}

var (
	// ErrFailedToDecode appears when failed to decode encoded data.
	ErrFailedToDecode = errors.New("failed to decode encoded data")
	// ErrEncoderIsNotInitialized appears when encoder is not initialized, run SetSecretKey first.
	ErrEncoderIsNotInitialized = errors.New("encoder is not initialized, run SetSecretKey first")
)

var _ controller.Encoder = (*AESGMCEncoder)(nil)

// Encode encodes byte.
func (s *AESGMCEncoder) Encode(byteToEncode []byte) ([]byte, error) {
	if !s.ready {
		return nil, ErrEncoderIsNotInitialized
	}
	encoded := s.encoder.Seal(nil, s.nonce, byteToEncode, nil)
	return encoded, nil
}

// Decode decodes bytes.
func (s *AESGMCEncoder) Decode(byteToDecode []byte) ([]byte, error) {
	if !s.ready {
		return nil, ErrEncoderIsNotInitialized
	}
	decoded, err := s.encoder.Open(nil, s.nonce, byteToDecode, nil) // расшифровываем
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, ErrFailedToDecode
	}
	return decoded, nil
}

// SetSecretKey sets secret key for encoder.
func (s *AESGMCEncoder) SetSecretKey(secretKey string) error {
	hashedKey := sha256.Sum256([]byte(secretKey))

	aesBlock, err := aes.NewCipher(hashedKey[:])
	if err != nil {
		return fmt.Errorf("an error occurred during secret encoder init: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return fmt.Errorf("an error occurred during secret encoder init: %w", err)
	}

	nonce := hashedKey[len(hashedKey)-aesgcm.NonceSize():]

	s.encoder = aesgcm
	s.nonce = nonce
	s.ready = true
	return nil
}
