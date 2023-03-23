package grpc_client

import (
	"context"

	pb "github.com/apolsh/yapr-gophkeeper/internal/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authConfig struct {
	token       string
	authMethods map[string]bool
}

func unaryAuthInterceptor(config *authConfig) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		if config.authMethods[method] && config.token != "" {
			newCtx := attachToken(ctx, config.token)
			return invoker(newCtx, method, req, reply, cc, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)

	}
}

func streamAuthInterceptor(config *authConfig) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if config.authMethods[method] && config.token != "" {
			return streamer(attachToken(ctx, config.token), desc, cc, method, opts...)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func attachToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, pb.AuthKey, token)
}
