package grpcclient

import (
	"context"
	"time"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier/grpcmapping"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const requestTimeout = 3 * time.Second

type Client struct {
	client supplierv1.SupplierServiceClient
}

var _ supplier.Gateway = (*Client)(nil)

func Dial(address string) (*Client, *grpc.ClientConn, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &Client{client: supplierv1.NewSupplierServiceClient(connection)}, connection, nil
}

func New(client supplierv1.SupplierServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) SearchOffers(ctx context.Context, request supplier.SearchRequest) ([]supplier.TripOption, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	response, err := c.client.SearchOffers(ctx, grpcmapping.SearchRequestToProto(request))
	if err != nil {
		return nil, err
	}
	return grpcmapping.TripOptionsFromProto(response.Items), nil
}

func (c *Client) QuoteRefund(ctx context.Context, request supplier.RefundQuoteRequest) (*supplier.RefundQuote, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	response, err := c.client.QuoteRefund(ctx, grpcmapping.RefundRequestToProto(request))
	if err != nil {
		return nil, err
	}
	return grpcmapping.RefundQuoteFromProto(response), nil
}

func (c *Client) ExecuteRefund(ctx context.Context, request supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	response, err := c.client.ExecuteRefund(ctx, grpcmapping.ExecuteRefundRequestToProto(request))
	if err != nil {
		return nil, err
	}
	return grpcmapping.ExecuteRefundResultFromProto(response), nil
}
