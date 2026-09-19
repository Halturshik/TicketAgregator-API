package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testRequestID = "2b0d9d2b-0194-4dc6-804b-ec13ab1c6b40"

func TestUnaryClientInterceptorPropagatesRequestID(t *testing.T) {
	var output bytes.Buffer
	setTestDefaultLogger(t, &output)

	ctx := requestid.WithContext(context.Background(), testRequestID)
	invoker := func(
		ctx context.Context,
		_ string,
		_, _ any,
		_ *grpc.ClientConn,
		_ ...grpc.CallOption,
	) error {
		outgoing, ok := metadata.FromOutgoingContext(ctx)
		if !ok || len(outgoing.Get(requestid.MetadataKey)) != 1 || outgoing.Get(requestid.MetadataKey)[0] != testRequestID {
			t.Fatalf("outgoing request id = %v", outgoing.Get(requestid.MetadataKey))
		}
		return status.Error(codes.Unavailable, "unavailable")
	}

	err := UnaryClientInterceptor()(
		ctx, "/supplier.v1.SupplierService/SearchOffers", nil, nil, nil, invoker,
	)
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("code = %s", status.Code(err))
	}
	entry := decodeLogEntry(t, output.Bytes())
	if entry["request_id"] != testRequestID || entry["grpc_direction"] != "client" || entry["grpc_code"] != "Unavailable" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestUnaryClientInterceptorReplacesRequestIDAndPreservesMetadata(t *testing.T) {
	var output bytes.Buffer
	setTestDefaultLogger(t, &output)

	ctx := requestid.WithContext(context.Background(), testRequestID)
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
		requestid.MetadataKey, "stale-request-id",
		"authorization", "Bearer token",
	))
	invoker := func(
		ctx context.Context,
		_ string,
		_, _ any,
		_ *grpc.ClientConn,
		_ ...grpc.CallOption,
	) error {
		outgoing, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("outgoing metadata is missing")
		}
		if values := outgoing.Get(requestid.MetadataKey); len(values) != 1 || values[0] != testRequestID {
			t.Fatalf("outgoing request id = %v", values)
		}
		if values := outgoing.Get("authorization"); len(values) != 1 || values[0] != "Bearer token" {
			t.Fatalf("authorization metadata = %v", values)
		}
		return nil
	}

	if err := UnaryClientInterceptor()(
		ctx, "/supplier.v1.SupplierService/SearchOffers", nil, nil, nil, invoker,
	); err != nil {
		t.Fatalf("interceptor error = %v", err)
	}
}

func TestUnaryServerInterceptorRejectsAmbiguousRequestID(t *testing.T) {
	var output bytes.Buffer
	setTestDefaultLogger(t, &output)

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(requestid.MetadataKey, testRequestID, requestid.MetadataKey, "duplicate"),
	)
	handler := func(ctx context.Context, _ any) (any, error) {
		id, ok := requestid.FromContext(ctx)
		if !ok || id == "" || id == testRequestID || id == "duplicate" {
			t.Fatalf("ambiguous incoming request id was accepted: %q, %t", id, ok)
		}
		return "ok", nil
	}

	if _, err := UnaryServerInterceptor()(
		ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/supplier.v1.SupplierService/SearchOffers"}, handler,
	); err != nil {
		t.Fatalf("interceptor error = %v", err)
	}
}

func TestUnaryServerInterceptorReadsRequestID(t *testing.T) {
	var output bytes.Buffer
	setTestDefaultLogger(t, &output)

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(requestid.MetadataKey, testRequestID),
	)
	handler := func(ctx context.Context, _ any) (any, error) {
		id, ok := requestid.FromContext(ctx)
		if !ok || id != testRequestID {
			t.Fatalf("request id in handler = %q, %t", id, ok)
		}
		return "ok", nil
	}

	response, err := UnaryServerInterceptor()(
		ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/supplier.v1.SupplierService/SearchOffers"}, handler,
	)
	if err != nil || response != "ok" {
		t.Fatalf("response = %v, error = %v", response, err)
	}
	entry := decodeLogEntry(t, output.Bytes())
	if entry["request_id"] != testRequestID || entry["grpc_direction"] != "server" || entry["grpc_code"] != "OK" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func setTestDefaultLogger(t *testing.T, output *bytes.Buffer) {
	t.Helper()
	log, err := New(Options{Service: "test", Format: FormatJSON, Level: "debug", Output: output})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	previous := slog.Default()
	slog.SetDefault(log)
	t.Cleanup(func() { slog.SetDefault(previous) })
}

func decodeLogEntry(t *testing.T, value []byte) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(value, &entry); err != nil {
		t.Fatalf("decode log: %v; output = %q", err, value)
	}
	return entry
}
