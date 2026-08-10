package search

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	searchv1 "github.com/christolx/cartlabs/internal/contract/searchv1"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeRepository struct {
	ids     []string
	applied bool
	err     error
}

func (f *fakeRepository) Search(context.Context, string, int) ([]string, error) { return f.ids, f.err }
func (f *fakeRepository) Upsert(context.Context, string, Document) (bool, error) {
	return f.applied, f.err
}
func (f *fakeRepository) Count(context.Context) (int64, error) { return 0, f.err }
func (f *fakeRepository) Prune(context.Context, []string, time.Time) (int64, error) {
	return 0, f.err
}

func TestServerValidatesAndReturnsCandidates(t *testing.T) {
	repository := &fakeRepository{ids: []string{"01989f00-0000-7000-8000-000000000201"}}
	server := NewServer(repository, NewMetrics(prometheus.NewRegistry()))
	response, err := server.SearchProducts(context.Background(), &searchv1.SearchProductsRequest{Query: " basket ", Limit: 20})
	if err != nil || len(response.ProductIds) != 1 {
		t.Fatalf("response=%v err=%v", response, err)
	}
	if _, err := server.SearchProducts(context.Background(), &searchv1.SearchProductsRequest{Query: "", Limit: 20}); status.Code(err) != 3 {
		t.Fatalf("invalid search error=%v", err)
	}
}

func TestServerMarksCandidateOverflow(t *testing.T) {
	repository := &fakeRepository{ids: []string{"one", "two", "three"}}
	server := NewServer(repository, NewMetrics(prometheus.NewRegistry()))
	response, err := server.SearchProducts(context.Background(), &searchv1.SearchProductsRequest{Query: "basket", Limit: 2})
	if err != nil || !response.Truncated || len(response.ProductIds) != 2 {
		t.Fatalf("response=%v err=%v", response, err)
	}
}

func TestServerUpsertValidation(t *testing.T) {
	repository := &fakeRepository{applied: true}
	server := NewServer(repository, NewMetrics(prometheus.NewRegistry()))
	request := &searchv1.UpsertProductRequest{EventId: "01989f00-0000-7000-8000-000000000901", Document: &searchv1.SearchDocument{
		ProductId: "01989f00-0000-7000-8000-000000000201", Name: "Market Basket", Description: "Woven",
		CategorySlug: "home-living", UpdatedAt: timestamppb.New(time.Now().UTC()),
	}}
	response, err := server.UpsertProduct(context.Background(), request)
	if err != nil || !response.Applied {
		t.Fatalf("response=%v err=%v", response, err)
	}
	request.Document.CategorySlug = "INVALID SLUG"
	if _, err := server.UpsertProduct(context.Background(), request); status.Code(err) != 3 {
		t.Fatalf("invalid upsert error=%v", err)
	}
}

func TestAuthInterceptor(t *testing.T) {
	interceptor := UnaryAuthInterceptor("01234567890123456789012345678901")
	handler := func(context.Context, any) (any, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/cartlabs.search.v1.SearchService/SearchProducts"}
	if _, err := interceptor(context.Background(), nil, info, handler); status.Code(err) != 16 {
		t.Fatalf("missing auth error=%v", err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer 01234567890123456789012345678901"))
	if result, err := interceptor(ctx, nil, info, handler); err != nil || result != "ok" {
		t.Fatalf("result=%v err=%v", result, err)
	}
}

func TestClientCircuitFailsFastAfterTransportFailure(t *testing.T) {
	client, err := NewClient("127.0.0.1:1", "01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.Search(context.Background(), "basket", 20); err == nil {
		t.Fatal("unreachable service returned no error")
	}
	started := time.Now()
	if _, err := client.Search(context.Background(), "basket", 20); err == nil {
		t.Fatal("open circuit returned no error")
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("open circuit elapsed=%s", elapsed)
	}
}

type fakeUpserter struct {
	eventID  string
	document Document
	err      error
}

func (f *fakeUpserter) Upsert(_ context.Context, eventID string, document Document) (bool, error) {
	f.eventID, f.document = eventID, document
	return true, f.err
}

func TestSyncHandlerRejectsInvalidAndPropagatesFailure(t *testing.T) {
	client := &fakeUpserter{}
	handler := NewSyncHandler(client)
	document := Document{ProductID: "01989f00-0000-7000-8000-000000000201", Name: "Basket", Description: "Woven", CategorySlug: "home", UpdatedAt: time.Now().UTC()}
	body, _ := json.Marshal(UpsertEvent{ID: "01989f00-0000-7000-8000-000000000901", Type: EventType, Data: document})
	if err := handler.Handle(context.Background(), body); err != nil || client.document.ProductID != document.ProductID {
		t.Fatalf("document=%#v err=%v", client.document, err)
	}
	if err := handler.Handle(context.Background(), []byte("{")); err == nil {
		t.Fatal("malformed event accepted")
	}
	client.err = errors.New("service unavailable")
	if err := handler.Handle(context.Background(), body); !errors.Is(err, client.err) {
		t.Fatalf("error=%v", err)
	}
}
