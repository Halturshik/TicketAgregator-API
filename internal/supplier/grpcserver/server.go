package grpcserver

import (
	"context"
	"errors"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier/grpcmapping"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	supplierv1.UnimplementedSupplierServiceServer
	service supplier.Service
}

func New(service supplier.Service) *Server {
	return &Server{service: service}
}

func (s *Server) SearchOffers(ctx context.Context, request *supplierv1.SearchOffersRequest) (*supplierv1.SearchOffersResponse, error) {
	items, err := s.service.SearchOffers(ctx, grpcmapping.SearchRequestFromProto(request))
	if err != nil {
		return nil, grpcError(err)
	}
	return &supplierv1.SearchOffersResponse{Items: grpcmapping.TripOptionsToProto(items)}, nil
}

func (s *Server) QuoteRefund(ctx context.Context, request *supplierv1.QuoteRefundRequest) (*supplierv1.QuoteRefundResponse, error) {
	quote, err := s.service.QuoteRefund(ctx, grpcmapping.RefundRequestFromProto(request))
	if err != nil {
		return nil, grpcError(err)
	}
	return grpcmapping.RefundQuoteToProto(quote), nil
}

func (s *Server) ExecuteRefund(ctx context.Context, request *supplierv1.ExecuteRefundRequest) (*supplierv1.ExecuteRefundResponse, error) {
	result, err := s.service.ExecuteRefund(ctx, grpcmapping.ExecuteRefundRequestFromProto(request))
	if err != nil {
		return nil, grpcError(err)
	}
	return grpcmapping.ExecuteRefundResultToProto(result), nil
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	case errors.Is(err, supplier.ErrInvalidSearchRequest), errors.Is(err, supplier.ErrInvalidRefundRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, supplier.ErrSupplierNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, supplier.ErrNoCarriers):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, supplier.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, supplier.ErrTicketAlreadyRefunded):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		logger.Error("Внутренняя ошибка gRPC-сервера поставщиков: %v", err)
		return status.Error(codes.Internal, "supplier internal error")
	}
}
