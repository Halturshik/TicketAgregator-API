package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		connection *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		options ...grpc.CallOption,
	) error {
		id, ok := requestid.FromContext(ctx)
		if !ok {
			id = requestid.New()
			ctx = requestid.WithContext(ctx, id)
		}
		outgoing, _ := metadata.FromOutgoingContext(ctx)
		outgoing = outgoing.Copy()
		if outgoing == nil {
			outgoing = metadata.MD{}
		}
		outgoing.Set(requestid.MetadataKey, id)
		ctx = metadata.NewOutgoingContext(ctx, outgoing)
		startedAt := time.Now()
		err := invoker(ctx, method, req, reply, connection, options...)
		if !isHealthMethod(method) {
			logGRPCCall(ctx, "client", method, startedAt, err)
		}
		return err
	}
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		id := requestid.New()
		if incoming, ok := metadata.FromIncomingContext(ctx); ok {
			values := incoming.Get(requestid.MetadataKey)
			if len(values) == 1 {
				id = requestid.Resolve(values[0])
			}
		}
		ctx = requestid.WithContext(ctx, id)
		_ = grpc.SetHeader(ctx, metadata.Pairs(requestid.MetadataKey, id))
		startedAt := time.Now()
		response, err := handler(ctx, req)
		if !isHealthMethod(info.FullMethod) {
			logGRPCCall(ctx, "server", info.FullMethod, startedAt, err)
		}
		return response, err
	}
}

func logGRPCCall(ctx context.Context, direction, method string, startedAt time.Time, err error) {
	code := status.Code(err)
	attrs := []slog.Attr{
		slog.String("grpc_direction", direction),
		slog.String("grpc_method", method),
		slog.String("grpc_code", code.String()),
		slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
	}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}
	slog.LogAttrs(ctx, grpcLogLevel(code), "gRPC-вызов завершён", attrs...)
}

func grpcLogLevel(code codes.Code) slog.Level {
	switch code {
	case codes.OK, codes.Canceled, codes.InvalidArgument, codes.NotFound,
		codes.AlreadyExists, codes.FailedPrecondition, codes.Unauthenticated,
		codes.PermissionDenied:
		return slog.LevelInfo
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Unavailable:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func isHealthMethod(method string) bool {
	return method == grpchealthv1.Health_Check_FullMethodName || method == grpchealthv1.Health_Watch_FullMethodName
}
