package search

import (
	"context"
	"crypto/subtle"
	"regexp"
	"strings"
	"time"

	searchv1 "github.com/christolx/cartlabs/internal/contract/searchv1"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var categorySlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Metrics struct {
	Requests *prometheus.CounterVec
	Latency  *prometheus.HistogramVec
	Upserts  *prometheus.CounterVec
}

func NewMetrics(registry prometheus.Registerer) *Metrics {
	metrics := &Metrics{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cartlabs_search_requests_total", Help: "Search requests by result."}, []string{"result"}),
		Latency:  prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "cartlabs_search_request_duration_seconds", Help: "Search request latency.", Buckets: prometheus.DefBuckets}, []string{"result"}),
		Upserts:  prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cartlabs_search_upserts_total", Help: "Search index upserts by result."}, []string{"result"}),
	}
	registry.MustRegister(metrics.Requests, metrics.Latency, metrics.Upserts)
	return metrics
}

type Server struct {
	searchv1.UnimplementedSearchServiceServer
	repository Repository
	metrics    *Metrics
}

func NewServer(repository Repository, metrics *Metrics) *Server {
	return &Server{repository: repository, metrics: metrics}
}

func (s *Server) SearchProducts(ctx context.Context, request *searchv1.SearchProductsRequest) (*searchv1.SearchProductsResponse, error) {
	started := time.Now()
	query := strings.TrimSpace(request.GetQuery())
	limit := int(request.GetLimit())
	if len(query) == 0 || len(query) > 100 || limit < 1 || limit > MaximumRPCItems {
		s.metrics.Requests.WithLabelValues("invalid").Inc()
		s.metrics.Latency.WithLabelValues("invalid").Observe(time.Since(started).Seconds())
		return nil, status.Error(codes.InvalidArgument, "query and limit are invalid")
	}
	ids, err := s.repository.Search(ctx, query, limit+1)
	result := "success"
	if err != nil {
		result = "failed"
	}
	s.metrics.Requests.WithLabelValues(result).Inc()
	s.metrics.Latency.WithLabelValues(result).Observe(time.Since(started).Seconds())
	if err != nil {
		return nil, status.Error(codes.Internal, "search unavailable")
	}
	truncated := len(ids) > limit
	if truncated {
		ids = ids[:limit]
	}
	return &searchv1.SearchProductsResponse{ProductIds: ids, Truncated: truncated}, nil
}

func (s *Server) UpsertProduct(ctx context.Context, request *searchv1.UpsertProductRequest) (*searchv1.UpsertProductResponse, error) {
	document := request.GetDocument()
	if document == nil || document.GetUpdatedAt() == nil || document.GetUpdatedAt().CheckValid() != nil ||
		!validUUID(request.GetEventId()) || !validUUID(document.GetProductId()) ||
		len(strings.TrimSpace(document.GetName())) < 2 || len(document.GetName()) > 160 ||
		len(document.GetDescription()) > 5000 || !categorySlugPattern.MatchString(document.GetCategorySlug()) {
		s.metrics.Upserts.WithLabelValues("invalid").Inc()
		return nil, status.Error(codes.InvalidArgument, "search document is invalid")
	}
	applied, err := s.repository.Upsert(ctx, request.GetEventId(), Document{
		ProductID: document.GetProductId(), Name: strings.TrimSpace(document.GetName()), Description: strings.TrimSpace(document.GetDescription()),
		CategorySlug: strings.TrimSpace(document.GetCategorySlug()), UpdatedAt: document.GetUpdatedAt().AsTime().UTC(),
	})
	result := "duplicate"
	if applied {
		result = "applied"
	}
	if err != nil {
		result = "failed"
	}
	s.metrics.Upserts.WithLabelValues(result).Inc()
	if err != nil {
		return nil, status.Error(codes.Internal, "index update failed")
	}
	return &searchv1.UpsertProductResponse{Applied: applied}, nil
}

func (s *Server) PruneProducts(ctx context.Context, request *searchv1.PruneProductsRequest) (*searchv1.PruneProductsResponse, error) {
	cutoff := request.GetSourceSnapshotStartedAt()
	if cutoff == nil || cutoff.CheckValid() != nil || len(request.GetKeepProductIds()) > 100000 {
		return nil, status.Error(codes.InvalidArgument, "prune snapshot is invalid")
	}
	for _, id := range request.GetKeepProductIds() {
		if !validUUID(id) {
			return nil, status.Error(codes.InvalidArgument, "prune product ID is invalid")
		}
	}
	deleted, err := s.repository.Prune(ctx, request.GetKeepProductIds(), cutoff.AsTime().UTC())
	if err != nil {
		return nil, status.Error(codes.Internal, "index prune failed")
	}
	return &searchv1.PruneProductsResponse{Deleted: uint64(deleted)}, nil
}

func validUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}

func UnaryAuthInterceptor(token string) grpc.UnaryServerInterceptor {
	want := []byte("Bearer " + token)
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") {
			return handler(ctx, request)
		}
		values := metadata.ValueFromIncomingContext(ctx, "authorization")
		if len(values) != 1 || len(values[0]) != len(want) || subtle.ConstantTimeCompare([]byte(values[0]), want) != 1 {
			return nil, status.Error(codes.Unauthenticated, "service token required")
		}
		return handler(ctx, request)
	}
}

func protoDocument(document Document) *searchv1.SearchDocument {
	return &searchv1.SearchDocument{ProductId: document.ProductID, Name: document.Name, Description: document.Description,
		CategorySlug: document.CategorySlug, UpdatedAt: timestamppb.New(document.UpdatedAt)}
}
