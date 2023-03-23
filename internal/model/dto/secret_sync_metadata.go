package dto

// SecretSyncMetadata synchronization metadata of secret items
type SecretSyncMetadata struct {
	// ID identifier of secret item
	ID string
	// Hash of secret item content
	Hash string
	// Timestamp of secret item last modification
	Timestamp int64
}
