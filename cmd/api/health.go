package main

import (
	"context"
	"database/sql"
	"fmt"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	platformhealth "github.com/Halturshik/TicketAgregator-API/internal/platform/health"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func newHealthHandler(
	db *sql.DB,
	redisClient *redis.Client,
	supplierConnection grpc.ClientConnInterface,
) *platformhealth.Handler {
	supplierHealthClient := grpchealthv1.NewHealthClient(supplierConnection)
	return platformhealth.New(
		platformhealth.DefaultTimeout,
		platformhealth.Dependency{
			Name:  "postgres",
			Probe: platformhealth.ProbeFunc(db.PingContext),
		},
		platformhealth.Dependency{
			Name: "redis",
			Probe: platformhealth.ProbeFunc(func(ctx context.Context) error {
				return redisClient.Ping(ctx).Err()
			}),
		},
		platformhealth.Dependency{
			Name: "supplier",
			Probe: platformhealth.ProbeFunc(func(ctx context.Context) error {
				response, err := supplierHealthClient.Check(ctx, &grpchealthv1.HealthCheckRequest{
					Service: supplierv1.SupplierService_ServiceDesc.ServiceName,
				})
				if err != nil {
					return err
				}
				if response.Status != grpchealthv1.HealthCheckResponse_SERVING {
					return fmt.Errorf("gRPC-сервис поставщиков не готов: status=%s", response.Status)
				}
				return nil
			}),
		},
	)
}
