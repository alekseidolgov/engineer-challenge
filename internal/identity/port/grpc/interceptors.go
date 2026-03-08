package grpc

import (
	"context"
	"log/slog"
	"time"

	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(logger *slog.Logger) grpclib.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpclib.UnaryServerInfo, handler grpclib.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		code := status.Code(err)
		attrs := []any{
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"code", code.String(),
		}

		if err != nil {
			logger.Error("gRPC request failed", append(attrs, "error", err.Error())...)
		} else {
			logger.Info("gRPC request", attrs...)
		}

		return resp, err
	}
}
