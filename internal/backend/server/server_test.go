package server

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/apolsh/yapr-gophkeeper/internal/backend/service"
	"github.com/apolsh/yapr-gophkeeper/internal/backend/storage"
	pb "github.com/apolsh/yapr-gophkeeper/internal/grpc/proto"
	"github.com/apolsh/yapr-gophkeeper/internal/mocks"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServerSuite struct {
	suite.Suite
	service      *mocks.MockGophkeeperService
	server       *GRPCGophkeeperServer
	tokenManager *mocks.MockTokenManager
	ctrl         *gomock.Controller
	client       pb.GophkeeperClient
}

func TestGRPCServer(t *testing.T) {
	suite.Run(t, new(GRPCServerSuite))
}

func (s *GRPCServerSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.ctrl = ctrl
	s.service = mocks.NewMockGophkeeperService(ctrl)
	s.tokenManager = mocks.NewMockTokenManager(ctrl)
	server := NewGRPCGophkeeperServer(":3333", s.service, s.tokenManager)
	s.server = server
	go func() {
		err := server.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()
	conn, err := grpc.Dial(":3333", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	s.client = pb.NewGophkeeperClient(conn)
}

func (s *GRPCServerSuite) TearDownTest() {
	if s.server != nil {
		err := s.server.Stop(context.Background())
		time.Sleep(500 * time.Millisecond) //delay between stop and start
		if err != nil {
			log.Fatal(err)
		}
	}
}

var (
	userID             int64 = 1
	userLogin                = "login"
	userPassword             = "password"
	userHashedPassword       = "hashedPassword"
	userTimestamp      int64 = 1679391035652
	userToken                = "token"
	user                     = model.User{ID: userID, Login: userLogin, HashedPassword: userHashedPassword, Timestamp: userTimestamp}
	secretID                 = "1"
	secretName               = "secretName"
	secretDescription        = "secretDescription"
	secretHash               = "hash1"
	secretTimestamp    int64 = 111
	secretType               = model.Credentials
	secretEncContent         = []byte("bytes")
	syncMetas                = []dto.SecretSyncMetadata{{ID: secretID, Hash: secretHash, Timestamp: secretTimestamp}}
	encodedSecret            = model.EncodedSecret{ID: secretID, Name: secretName, Owner: userID, Description: secretDescription, Type: secretType, EncodedContent: secretEncContent, Hash: secretHash, Timestamp: secretTimestamp}
)

func (s *GRPCServerSuite) TestRegisterSuccess() {
	s.service.EXPECT().Register(gomock.All(), userLogin, userPassword).Return(userToken, user, nil)
	authMeta, err := s.client.Register(context.Background(), &pb.Credentials{Login: userLogin, Password: userPassword})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), authMeta)
	assert.NotNil(s.T(), authMeta.GetUser())
	assert.Equal(s.T(), authMeta.GetUser().GetID(), userID)
	assert.Equal(s.T(), authMeta.GetUser().GetUsername(), userLogin)
	assert.Equal(s.T(), authMeta.GetUser().GetPasswordHash(), userHashedPassword)
	assert.Equal(s.T(), authMeta.GetUser().GetTimestamp(), userTimestamp)
	assert.Equal(s.T(), authMeta.GetToken(), userToken)
}

func (s *GRPCServerSuite) TestRegisterErrorLoginIsAlreadyUsed() {
	s.service.EXPECT().Register(gomock.All(), userLogin, userPassword).Return("", model.User{}, storage.ErrorLoginIsAlreadyUsed)
	_, err := s.client.Register(context.Background(), &pb.Credentials{Login: userLogin, Password: userPassword})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.AlreadyExists, st.Code())
	assert.Equal(s.T(), storage.ErrorLoginIsAlreadyUsed.Error(), st.Message())
}

func (s *GRPCServerSuite) TestRegisterErrorEmptyValue() {
	s.service.EXPECT().Register(gomock.All(), userLogin, userPassword).Return("", model.User{}, service.ErrorEmptyValue)
	_, err := s.client.Register(context.Background(), &pb.Credentials{Login: userLogin, Password: userPassword})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.Unauthenticated, st.Code())
	assert.Equal(s.T(), service.ErrorEmptyValue.Error(), st.Message())
}

func (s *GRPCServerSuite) TestLoginSuccess() {
	s.service.EXPECT().Login(gomock.All(), userLogin, userPassword).Return(userToken, user, nil)
	authMeta, err := s.client.Login(context.Background(), &pb.Credentials{Login: userLogin, Password: userPassword})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), authMeta)
	assert.NotNil(s.T(), authMeta.GetUser())
	assert.Equal(s.T(), authMeta.GetUser().GetID(), userID)
	assert.Equal(s.T(), authMeta.GetUser().GetUsername(), userLogin)
	assert.Equal(s.T(), authMeta.GetUser().GetPasswordHash(), userHashedPassword)
	assert.Equal(s.T(), authMeta.GetUser().GetTimestamp(), userTimestamp)
	assert.Equal(s.T(), authMeta.GetToken(), userToken)
}

func (s *GRPCServerSuite) TestLoginErrorEmptyValue() {
	s.service.EXPECT().Login(gomock.All(), userLogin, userPassword).Return("", model.User{}, service.ErrorEmptyValue)
	_, err := s.client.Login(context.Background(), &pb.Credentials{Login: userLogin, Password: userPassword})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.Unauthenticated, st.Code())
	assert.Equal(s.T(), service.ErrorEmptyValue.Error(), st.Message())
}

func (s *GRPCServerSuite) TestGetSecretSyncMetaSuccess() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().GetSecretSyncMetaByUser(gomock.Any(), userID).Return(syncMetas, nil)
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	meta, err := s.client.GetSecretSyncMeta(ctx, &emptypb.Empty{})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(meta.GetItems()))
	item := meta.GetItems()[0]
	assert.Equal(s.T(), secretID, item.SecretID)
	assert.Equal(s.T(), secretHash, item.Hash)
	assert.Equal(s.T(), secretTimestamp, item.Timestamp)
}

func (s *GRPCServerSuite) TestGetSecretSyncMetaNoAuth() {
	_, err := s.client.GetSecretSyncMeta(context.Background(), &emptypb.Empty{})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.Unauthenticated, st.Code())
}

func (s *GRPCServerSuite) TestGetSecretSyncMetaError() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().GetSecretSyncMetaByUser(gomock.Any(), userID).Return([]dto.SecretSyncMetadata{}, errors.New("some error"))
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.GetSecretSyncMeta(ctx, &emptypb.Empty{})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.Unknown, st.Code())
}

func (s *GRPCServerSuite) TestGetSecretSuccess() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().GetSecret(gomock.Any(), int(userID), secretID).Return(encodedSecret, nil)
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	secret, err := s.client.GetSecret(ctx, &pb.SecretID{SecretID: secretID})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), secretID, secret.GetId())
	assert.Equal(s.T(), secretName, secret.GetName())
	assert.Equal(s.T(), secretDescription, secret.GetDescription())
	assert.Equal(s.T(), userID, secret.GetOwner())
	assert.Equal(s.T(), strings.ToLower(secretType), strings.ToLower(secret.GetType().String()))
	assert.Equal(s.T(), secretEncContent, secret.GetEncData())
	assert.Equal(s.T(), secretHash, secret.GetHash())
	assert.Equal(s.T(), secretTimestamp, secret.GetDateLastModified())
}

func (s *GRPCServerSuite) TestGetSecretErrOwnerMissmatch() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().GetSecret(gomock.Any(), int(userID), secretID).Return(model.EncodedSecret{}, service.ErrOwnerMissmatch)
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.GetSecret(ctx, &pb.SecretID{SecretID: secretID})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.PermissionDenied, st.Code())
	assert.Equal(s.T(), service.ErrOwnerMissmatch.Error(), st.Message())
}

func (s *GRPCServerSuite) TestGetSecretErr() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().GetSecret(gomock.Any(), int(userID), secretID).Return(model.EncodedSecret{}, errors.New(""))
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.GetSecret(ctx, &pb.SecretID{SecretID: secretID})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.Unknown, st.Code())
}

func (s *GRPCServerSuite) TestSaveEncodedSecretSuccess() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().SaveEncodedSecret(gomock.Any(), int(userID), encodedSecret).Return(nil)
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.SaveEncodedSecret(ctx, pb.EncSecretProtoFromEncSecret(encodedSecret))
	assert.NoError(s.T(), err)
}

func (s *GRPCServerSuite) TestSaveEncodedSecretErrOwnerMissmatch() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().SaveEncodedSecret(gomock.Any(), int(userID), encodedSecret).Return(service.ErrOwnerMissmatch)
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.SaveEncodedSecret(ctx, pb.EncSecretProtoFromEncSecret(encodedSecret))
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.PermissionDenied, st.Code())
	assert.Equal(s.T(), service.ErrOwnerMissmatch.Error(), st.Message())
}

func (s *GRPCServerSuite) TestSaveEncodedSecretErr() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().SaveEncodedSecret(gomock.Any(), int(userID), encodedSecret).Return(errors.New(""))
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.SaveEncodedSecret(ctx, pb.EncSecretProtoFromEncSecret(encodedSecret))
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.Unknown, st.Code())
}

func (s *GRPCServerSuite) TestDeleteSecretSuccess() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().DeleteSecret(gomock.Any(), int(userID), secretID).Return(nil)
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.DeleteSecret(ctx, &pb.SecretID{SecretID: secretID})
	assert.NoError(s.T(), err)
}

func (s *GRPCServerSuite) TestDeleteSecretErrOwnerMissmatch() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().DeleteSecret(gomock.Any(), int(userID), secretID).Return(service.ErrOwnerMissmatch)
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.DeleteSecret(ctx, &pb.SecretID{SecretID: secretID})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.PermissionDenied, st.Code())
	assert.Equal(s.T(), service.ErrOwnerMissmatch.Error(), st.Message())
}

func (s *GRPCServerSuite) TestDeleteSecretError() {
	s.tokenManager.EXPECT().ParseToken(userToken).Return(userID, nil)
	s.service.EXPECT().DeleteSecret(gomock.Any(), int(userID), secretID).Return(errors.New(""))
	ctx := metadata.AppendToOutgoingContext(context.Background(), pb.AuthKey, userToken)
	_, err := s.client.DeleteSecret(ctx, &pb.SecretID{SecretID: secretID})
	assert.NotNil(s.T(), err)
	st, ok := status.FromError(err)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), codes.Unknown, st.Code())
}
