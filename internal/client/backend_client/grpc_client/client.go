package grpc_client

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophkeeper/internal/client/controller"
	pb "github.com/apolsh/yapr-gophkeeper/internal/grpc/proto"
	"github.com/apolsh/yapr-gophkeeper/internal/logger"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	errs "github.com/apolsh/yapr-gophkeeper/internal/model/app_errors"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GophkeeperGRPCClient client for grpc interactions with backend.
type GophkeeperGRPCClient struct {
	client     pb.GophkeeperClient
	authConfig *authConfig
}

var _ controller.BackendClient = (*GophkeeperGRPCClient)(nil)

var log = logger.LoggerOfComponent("grpc_client")

// NewGophkeeperGRPCClient GophkeeperGRPCClient constructor.
func NewGophkeeperGRPCClient(serverURL string) *GophkeeperGRPCClient {
	authConfig := &authConfig{token: "", authMethods: pb.DefaultAuthMethods}

	conn, err := grpc.Dial(
		serverURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryAuthInterceptor(authConfig)),
		grpc.WithStreamInterceptor(streamAuthInterceptor(authConfig)),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewGophkeeperClient(conn)

	return &GophkeeperGRPCClient{client: client, authConfig: authConfig}
}

// NewGophkeeperGRPCClientTLS GophkeeperGRPCClient constructor with TLS.
func NewGophkeeperGRPCClientTLS(serverURL string, creds credentials.TransportCredentials) *GophkeeperGRPCClient {
	authConfig := &authConfig{token: "", authMethods: pb.DefaultAuthMethods}

	conn, err := grpc.Dial(
		serverURL,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(unaryAuthInterceptor(authConfig)),
		grpc.WithStreamInterceptor(streamAuthInterceptor(authConfig)),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewGophkeeperClient(conn)

	return &GophkeeperGRPCClient{client: client, authConfig: authConfig}
}

// SetAuthTokenForRequests sets authorization token for this client (add it to every required auth request).
func (c *GophkeeperGRPCClient) SetAuthTokenForRequests(token string) {
	c.authConfig.token = token
}

// Login login user.
func (c *GophkeeperGRPCClient) Login(ctx context.Context, login, password string) (string, model.User, error) {
	authMeta, err := c.client.Login(ctx, &pb.Credentials{Login: login, Password: password})
	if err != nil {
		log.Error(err)
		return "", model.User{}, handleStatusError(err)
	}
	user := pb.NewUserFromProtoUser(authMeta.GetUser())
	return authMeta.GetToken(), user, nil
}

// Register registers user.
func (c *GophkeeperGRPCClient) Register(ctx context.Context, login, password string) (string, model.User, error) {
	authMeta, err := c.client.Register(ctx, &pb.Credentials{Login: login, Password: password})
	if err != nil {
		log.Error(err)
		return "", model.User{}, handleStatusError(err)
	}
	user := pb.NewUserFromProtoUser(authMeta.GetUser())
	return authMeta.GetToken(), user, nil
}

// GetSecretSyncMeta returns metadata for synchronization metadata.
func (c *GophkeeperGRPCClient) GetSecretSyncMeta(ctx context.Context) ([]dto.SecretSyncMetadata, error) {
	res, err := c.client.GetSecretSyncMeta(ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err)
		return nil, handleStatusError(err)
	}
	sycMetas := make([]dto.SecretSyncMetadata, 0, len(res.Items))
	for _, protoSyncMeta := range res.Items {
		sycMetas = append(sycMetas, pb.SecretSyncMetadataFromProto(protoSyncMeta))
	}
	return sycMetas, nil
}

// GetSecretSyncMetaByName returns metadata for synchronization metadata for current secret by its name.
func (c *GophkeeperGRPCClient) GetSecretSyncMetaByName(ctx context.Context, name string) (dto.SecretSyncMetadata, error) {
	res, err := c.client.GetSecretSyncMetaByName(ctx, &pb.Name{Name: name})
	if err != nil {
		log.Error(err)
		return dto.SecretSyncMetadata{}, handleStatusError(err)
	}
	return pb.SecretSyncMetadataFromProto(res), nil
}

// GetSecretByID returns EncodedSecret by ID.
func (c *GophkeeperGRPCClient) GetSecretByID(ctx context.Context, id string) (model.EncodedSecret, error) {
	secret, err := c.client.GetSecret(ctx, &pb.SecretID{SecretID: id})
	if err != nil {
		log.Error(err)
		return model.EncodedSecret{}, handleStatusError(err)
	}
	return pb.EncodedSecretFromProto(secret), nil
}

// SaveEncodedSecret saves EncodedSecret.
func (c *GophkeeperGRPCClient) SaveEncodedSecret(ctx context.Context, encSecret model.EncodedSecret) error {
	_, err := c.client.SaveEncodedSecret(ctx, pb.EncSecretProtoFromEncSecret(encSecret))
	if err != nil {
		log.Error(err)
		return handleStatusError(err)
	}
	return nil
}

// DeleteSecret deletes EncodedSecret by ID.
func (c *GophkeeperGRPCClient) DeleteSecret(ctx context.Context, id string) error {
	_, err := c.client.DeleteSecret(ctx, &pb.SecretID{SecretID: id})
	if err != nil {
		log.Error(err)
		return handleStatusError(err)
	}
	return nil
}

func handleStatusError(err error) error {
	if s, ok := status.FromError(err); ok {
		if s.Code() == codes.Unauthenticated || s.Code() == codes.AlreadyExists {
			return errors.New(s.Message())
		}
		if s.Code() == codes.Unavailable {
			return errs.ErrServerIsNotAvailable
		}
	}
	return err
}
