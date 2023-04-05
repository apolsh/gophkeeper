package proto

import (
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
)

// AuthKey authorization key for grpc context.
const AuthKey = "authorization"

const servicePath = "/proto.Gophkeeper/"

// DefaultAuthMethods default authorization schema for grpc methods.
var DefaultAuthMethods = map[string]bool{
	servicePath + "Login":                   false,
	servicePath + "Register":                false,
	servicePath + "GetSecretSyncMeta":       true,
	servicePath + "GetSecretSyncMetaByName": true,
	servicePath + "GetSecret":               true,
	servicePath + "SaveEncodedSecret":       true,
	servicePath + "DeleteSecret":            true,
}

// NewUserFromProtoUser convert proto user to model user.
func NewUserFromProtoUser(protoUser *User) model.User {
	return model.User{
		ID:             protoUser.GetID(),
		Login:          protoUser.GetUsername(),
		HashedPassword: protoUser.GetPasswordHash(),
		Timestamp:      protoUser.GetTimestamp(),
	}
}

// NewProtoUserFromUser  convert model user to proto user.
func NewProtoUserFromUser(user model.User) *User {
	return &User{
		ID:           user.ID,
		Username:     user.Login,
		PasswordHash: user.HashedPassword,
		Timestamp:    user.Timestamp,
	}
}

// NewProtoSyncMetaFromSycMeta convert model syncMeta to proto syncMeta.
func NewProtoSyncMetaFromSycMeta(syncMeta dto.SecretSyncMetadata) *SecretSyncData {
	return &SecretSyncData{
		SecretID:  syncMeta.ID,
		Hash:      syncMeta.Hash,
		Timestamp: syncMeta.Timestamp,
	}
}

// EncodedSecretFromProto convert proto syncMeta to model syncMeta.
func EncodedSecretFromProto(proto *EncodedSecret) model.EncodedSecret {
	return model.EncodedSecret{
		ID:             proto.GetId(),
		Name:           proto.GetName(),
		Owner:          proto.GetOwner(),
		Description:    proto.GetDescription(),
		Type:           getTypeFromProto(proto.GetType()),
		EncodedContent: proto.GetEncData(),
		Hash:           proto.GetHash(),
		Timestamp:      proto.GetDateLastModified(),
	}
}

// EncSecretProtoFromEncSecret convert model EncodedSecret to proto.
func EncSecretProtoFromEncSecret(encSecret model.EncodedSecret) *EncodedSecret {
	return &EncodedSecret{
		Id:               encSecret.ID,
		Name:             encSecret.Name,
		Owner:            encSecret.Owner,
		Description:      encSecret.Description,
		Type:             getProtoSecretType(encSecret.Type),
		EncData:          encSecret.EncodedContent,
		Hash:             encSecret.Hash,
		DateLastModified: encSecret.Timestamp,
	}
}

// SecretSyncMetadataFromProto convert proto EncodedSecret to model.
func SecretSyncMetadataFromProto(proto *SecretSyncData) dto.SecretSyncMetadata {
	return dto.SecretSyncMetadata{
		ID:        proto.SecretID,
		Hash:      proto.Hash,
		Timestamp: proto.Timestamp,
	}
}

func getTypeFromProto(proto SECRET_TYPE) string {
	switch proto {
	case SECRET_TYPE_CREDENTIALS:
		return model.Credentials
	case SECRET_TYPE_TEXT:
		return model.Text
	case SECRET_TYPE_BINARY:
		return model.Binary
	case SECRET_TYPE_CARD:
		return model.Card
	default:
		return model.Text
	}
}

func getProtoSecretType(secretType string) SECRET_TYPE {
	switch secretType {
	case model.Credentials:
		return SECRET_TYPE_CREDENTIALS
	case model.Text:
		return SECRET_TYPE_TEXT
	case model.Binary:
		return SECRET_TYPE_BINARY
	case model.Card:
		return SECRET_TYPE_CARD
	default:
		return SECRET_TYPE_TEXT
	}
}
