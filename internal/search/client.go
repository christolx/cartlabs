package search

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	searchv1 "github.com/christolx/cartlabs/internal/contract/searchv1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	connection       *grpc.ClientConn
	client           searchv1.SearchServiceClient
	token            string
	timeout          time.Duration
	unavailableUntil atomic.Int64
}

func NewClient(address, token string) (*Client, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		return nil, fmt.Errorf("configure search client: %w", err)
	}
	return &Client{connection: connection, client: searchv1.NewSearchServiceClient(connection), token: token, timeout: 750 * time.Millisecond}, nil
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]string, error) {
	now := time.Now()
	if until := c.unavailableUntil.Load(); until > now.UnixNano() {
		return nil, fmt.Errorf("search circuit open")
	}
	ctx, cancel := context.WithTimeout(c.authorize(ctx), c.timeout)
	defer cancel()
	response, err := c.client.SearchProducts(ctx, &searchv1.SearchProductsRequest{Query: query, Limit: uint32(limit)})
	if err != nil {
		c.unavailableUntil.Store(now.Add(2 * time.Second).UnixNano())
		return nil, fmt.Errorf("search products: %w", err)
	}
	c.unavailableUntil.Store(0)
	if response.GetTruncated() {
		return nil, fmt.Errorf("search candidate result exceeded safe limit")
	}
	return response.GetProductIds(), nil
}

func (c *Client) Upsert(ctx context.Context, eventID string, document Document) (bool, error) {
	ctx, cancel := context.WithTimeout(c.authorize(ctx), c.timeout)
	defer cancel()
	response, err := c.client.UpsertProduct(ctx, &searchv1.UpsertProductRequest{EventId: eventID, Document: protoDocument(document)})
	if err != nil {
		return false, fmt.Errorf("upsert search product: %w", err)
	}
	return response.GetApplied(), nil
}

func (c *Client) Prune(ctx context.Context, keepProductIDs []string, snapshotStartedAt time.Time) (uint64, error) {
	ctx, cancel := context.WithTimeout(c.authorize(ctx), 10*time.Second)
	defer cancel()
	response, err := c.client.PruneProducts(ctx, &searchv1.PruneProductsRequest{
		KeepProductIds: keepProductIDs, SourceSnapshotStartedAt: timestamppb.New(snapshotStartedAt),
	})
	if err != nil {
		return 0, fmt.Errorf("prune search products: %w", err)
	}
	return response.GetDeleted(), nil
}

func (c *Client) authorize(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}

func (c *Client) Close() error { return c.connection.Close() }
