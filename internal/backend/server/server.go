//go:generate mockgen -destination=../../mocks/backend_gophkeeper_server.go -package=mocks github.com/apolsh/yapr-gophkeeper/internal/backend/server GophkeeperService
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strconv"

	"github.com/apolsh/yapr-gophkeeper/internal/backend/service"
	tokenManager "github.com/apolsh/yapr-gophkeeper/internal/backend/token_manager"
	pb "github.com/apolsh/yapr-gophkeeper/internal/grpc/proto"
	"github.com/apolsh/yapr-gophkeeper/internal/logger"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	errs "github.com/apolsh/yapr-gophkeeper/internal/model/app_errors"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var log = logger.LoggerOfComponent("grpc-handler")

type GophkeeperService interface {
	Login(ctx context.Context, login string, password string) (string, model.User, error)
	Register(ctx context.Context, login string, password string) (string, model.User, error)
	GetSecretSyncMetaByUser(ctx context.Context, id int64) ([]dto.SecretSyncMetadata, error)
	GetSecretSyncMetaByOwnerAndName(ctx context.Context, userID int, name string) (dto.SecretSyncMetadata, error)
	GetSecret(ctx context.Context, userID int, secretID string) (model.EncodedSecret, error)
	SaveEncodedSecret(ctx context.Context, ownerID int, secret model.EncodedSecret) error
	DeleteSecret(ctx context.Context, ownerID int, secretID string) error
}

type gophkeeperGRPCHandler struct {
	pb.UnimplementedGophkeeperServer
	service GophkeeperService
}

var empty emptypb.Empty

// GRPCGophkeeperServer grpc gophkeeper server.
type GRPCGophkeeperServer struct {
	addr         string
	service      GophkeeperService
	Server       *grpc.Server
	tokenManager tokenManager.TokenManager
	authMethods  map[string]bool
}

// Start starts server.
func (s *GRPCGophkeeperServer) Start() error {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryAuthInterceptor(s.tokenManager, s.authMethods)),
		grpc.StreamInterceptor(streamAuthInterceptor(s.tokenManager, s.authMethods)))

	s.Server = grpcServer

	pb.RegisterGophkeeperServer(grpcServer, &gophkeeperGRPCHandler{service: s.service})

	return grpcServer.Serve(listen)
}

// StartTLS запускает сервер с TLS шифрованием.
func (s *GRPCGophkeeperServer) StartTLS(cfg *tls.Config) error {
	tlsCreds := credentials.NewTLS(cfg)

	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryAuthInterceptor(s.tokenManager, s.authMethods)),
		grpc.StreamInterceptor(streamAuthInterceptor(s.tokenManager, s.authMethods)),
		grpc.Creds(tlsCreds),
	)

	s.Server = grpcServer

	pb.RegisterGophkeeperServer(grpcServer, &gophkeeperGRPCHandler{service: s.service})

	return grpcServer.Serve(listen)
}

// Stop stops server.
func (s *GRPCGophkeeperServer) Stop(_ context.Context) error {
	s.Server.GracefulStop()
	return nil
}

// NewGRPCGophkeeperServer GRPCGophkeeperServer constructor.
func NewGRPCGophkeeperServer(serverAddr string, service GophkeeperService, manager tokenManager.TokenManager) *GRPCGophkeeperServer {
	return &GRPCGophkeeperServer{
		addr:         serverAddr,
		service:      service,
		tokenManager: manager,
		authMethods:  pb.DefaultAuthMethods,
	}
}

// Login login user.
func (s *gophkeeperGRPCHandler) Login(ctx context.Context, credentials *pb.Credentials) (*pb.AuthMeta, error) {
	token, user, err := s.service.Login(ctx, credentials.Login, credentials.Password)
	if err != nil {
		log.Error(err)
		if errors.Is(errs.ErrorInvalidPassword, err) || errors.Is(errs.ErrorEmptyValue, err) || errors.Is(service.ErrUserNotFound, err) {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		return nil, err
	}
	return &pb.AuthMeta{Token: token, User: pb.NewProtoUserFromUser(user)}, nil
}

// Register register user.
func (s *gophkeeperGRPCHandler) Register(ctx context.Context, credentials *pb.Credentials) (*pb.AuthMeta, error) {
	token, user, err := s.service.Register(ctx, credentials.Login, credentials.Password)
	if err != nil {
		log.Error(err)
		if errors.Is(errs.ErrorEmptyValue, err) {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		if errors.Is(errs.ErrorLoginIsAlreadyUsed, err) {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
	}
	return &pb.AuthMeta{Token: token, User: pb.NewProtoUserFromUser(user)}, nil
}

// GetSecretSyncMeta returns metadata for secret synchronization.
func (s *gophkeeperGRPCHandler) GetSecretSyncMeta(ctx context.Context, _ *emptypb.Empty) (*pb.GetSecretsSyncDataResponse, error) {
	id, err := getUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to extract user id: "+err.Error())
	}

	syncMeta, err := s.service.GetSecretSyncMetaByUser(ctx, int64(id))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, err.Error())
	}
	protoSyncMeta := convertSecretSyncMetaToProto(syncMeta)

	return &pb.GetSecretsSyncDataResponse{Items: protoSyncMeta}, nil
}

func (s *gophkeeperGRPCHandler) GetSecretSyncMetaByName(ctx context.Context, name *pb.Name) (*pb.SecretSyncData, error) {
	id, err := getUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to extract user id: "+err.Error())
	}
	syncMeta, err := s.service.GetSecretSyncMetaByOwnerAndName(ctx, id, name.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Unknown, err.Error())
	}
	return pb.NewProtoSyncMetaFromSycMeta(syncMeta), nil
}

// GetSecret returns EncodedSecret by ID.
func (s *gophkeeperGRPCHandler) GetSecret(ctx context.Context, secretID *pb.SecretID) (*pb.EncodedSecret, error) {
	ownerID, err := getUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to extract user id: "+err.Error())
	}

	encodedSecret, err := s.service.GetSecret(ctx, ownerID, secretID.GetSecretID())
	if err != nil {
		if errors.Is(service.ErrOwnerMissmatch, err) {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	return pb.EncSecretProtoFromEncSecret(encodedSecret), nil
}

// SaveEncodedSecret saves EncodedSecret.
func (s *gophkeeperGRPCHandler) SaveEncodedSecret(ctx context.Context, encodedSecret *pb.EncodedSecret) (*emptypb.Empty, error) {
	ownerID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	encSecret := pb.EncodedSecretFromProto(encodedSecret)
	err = s.service.SaveEncodedSecret(ctx, ownerID, encSecret)
	if err != nil {
		if errors.Is(service.ErrOwnerMissmatch, err) {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	return &empty, err
}

// DeleteSecret delete EncodedSecret by ID.
func (s *gophkeeperGRPCHandler) DeleteSecret(ctx context.Context, secretID *pb.SecretID) (*emptypb.Empty, error) {
	ownerID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = s.service.DeleteSecret(ctx, ownerID, secretID.SecretID)
	if err != nil {
		if errors.Is(service.ErrOwnerMissmatch, err) {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func getUserID(ctx context.Context) (int, error) {
	meta, ok := metadata.FromIncomingContext(ctx)

	var value string

	if ok {
		values := meta.Get(UserIDKey)
		if len(values) > 0 {
			value = values[0]
		}
	}
	return strconv.Atoi(value)
}

func convertSecretSyncMetaToProto(syncMetas []dto.SecretSyncMetadata) []*pb.SecretSyncData {
	protoSyncMeta := make([]*pb.SecretSyncData, 0, len(syncMetas))
	for _, syncMeta := range syncMetas {
		protoSyncMeta = append(protoSyncMeta, pb.NewProtoSyncMetaFromSycMeta(syncMeta))
	}
	return protoSyncMeta
}
