package server

import (
	"context"
	"strconv"

	token "github.com/apolsh/yapr-gophkeeper/internal/backend/token_manager"
	pb "github.com/apolsh/yapr-gophkeeper/internal/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const UserIDKey = "user_id"

func unaryAuthInterceptor(tokenManager token.TokenManager, authMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if authMethods[info.FullMethod] {
			if newCtx, err := authorize(tokenManager, ctx); err != nil {
				return nil, err
			} else {
				return handler(newCtx, req)
			}

		}
		return handler(ctx, req)
	}
}

func streamAuthInterceptor(tokenManager token.TokenManager, authMethods map[string]bool) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if authMethods[info.FullMethod] {
			if newCtx, err := authorize(tokenManager, ss.Context()); err != nil {
				return err
			} else {
				sw := newStreamContextWrapper(ss)
				sw.SetContext(newCtx)
				return handler(srv, sw)
			}
		}
		return handler(srv, ss)
	}
}

func authorize(tokenManager token.TokenManager, ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md[pb.AuthKey]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not set")
	}

	accessToken := values[0]
	userID, err := tokenManager.ParseToken(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token is invalid: %v", err)
	}

	md.Append(UserIDKey, strconv.Itoa(int(userID)))
	return metadata.NewIncomingContext(ctx, md), nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the context for this stream.
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// SetContext sets context to this stream.
func (w *wrappedStream) SetContext(ctx context.Context) {
	w.ctx = ctx
}

// RecvMsg blocks until it receives a message into m or the stream is done.
func (w *wrappedStream) RecvMsg(m interface{}) error {
	return w.ServerStream.RecvMsg(m)
}

// SendMsg sends a message. On error, SendMsg aborts the stream and the error is returned directly.
func (w *wrappedStream) SendMsg(m interface{}) error {
	return w.ServerStream.SendMsg(m)
}

// StreamContextWrapper stream wrapper for rewrite the context.
type StreamContextWrapper interface {
	grpc.ServerStream
	SetContext(context.Context)
}

func newStreamContextWrapper(ss grpc.ServerStream) StreamContextWrapper {
	ctx := ss.Context()
	return &wrappedStream{
		ss,
		ctx,
	}
}
