package model

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

var _ SecretItem = (*BinarySecretItem)(nil)

// BinarySecretItem binary implementation of SecretItem.
type BinarySecretItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SecretType  string `json:"secretType"`
	Binary      []byte `json:"binary"`
	Filename    string `json:"filename"`
	outputPath  string
}

// GetSecretPayload returns text implementation of secret item payload.
func (c *BinarySecretItem) GetSecretPayload() string {
	var output string
	var outputDirPath string
	if c.outputPath != "" {
		outputDirPath = c.outputPath
	} else {
		outputDirPath, _ = os.UserHomeDir()
	}
	absFilepath := filepath.Join(outputDirPath, c.Filename)
	err := os.WriteFile(absFilepath, c.Binary, 0755)
	if err != nil {
		output = fmt.Sprintf("failed to save fail: %s", err.Error())
	} else {
		output = fmt.Sprintf("file saved to: %s", absFilepath)
	}
	return fmt.Sprintf("[Binary]: %s \n", output)
}

// NewEncodedSecret encodes secret item.
func (c *BinarySecretItem) NewEncodedSecret(encodeFunction func(byteToDecode []byte) ([]byte, error), ownerID int64) (EncodedSecret, error) {
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
		Type:           Binary,
		EncodedContent: encoded,
		Hash:           base64.StdEncoding.EncodeToString(hash[:]),
		Timestamp:      time.Now().UTC().UnixMilli(),
	}
	return encodedSecret, nil
}

// GetType returns secret item type.
func (c *BinarySecretItem) GetType() string {
	return c.SecretType
}

// DecodeBinarySecretItem decodes EncodedSecret item into BinarySecretItem.
func DecodeBinarySecretItem(decode func(byteToEncode []byte) ([]byte, error), encoded EncodedSecret) (*BinarySecretItem, error) {
	var binarySecret BinarySecretItem
	decodedBytes, err := decode(encoded.EncodedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode binary secret: %w", err)
	}
	err = json.Unmarshal(decodedBytes, &binarySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse binary secret: %w", err)
	}
	return &binarySecret, nil
}

// NewBinarySecretItem BinarySecretItem constructor.
func NewBinarySecretItem(name, description, path string) (*BinarySecretItem, error) {
	stats, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file specified by path does not exist")
		}
		return nil, fmt.Errorf("failed to get information about specified file: %w", err)
	}
	if stats.IsDir() {
		return nil, fmt.Errorf("a directory was passed, you must specify the path to the FILE")
	}

	filename := filepath.Base(path)
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read specified file: %w", err)
	}

	return &BinarySecretItem{
		Name:        name,
		Description: description,
		SecretType:  Binary,
		Binary:      fileBytes,
		Filename:    filename,
	}, nil
}

// SetOutputPath sets output path, uses for decode binary file.
func (c *BinarySecretItem) SetOutputPath(path string) error {
	err := isCorrectDirectoryPath(path)
	if err != nil {
		return err
	}
	c.outputPath = path
	return nil
}

func isCorrectDirectoryPath(path string) error {
	stats, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory specified by path does not exist")
		}
		return err
	}
	if !stats.IsDir() {
		return fmt.Errorf("specified path doest not lead to a directory")
	}
	return nil
}
