package observability

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/christolx/cartlabs"

type TraceConfig struct {
	ServiceName string
	Environment string
	Endpoint    string
	SampleRatio float64
}

func SetupTracing(ctx context.Context, cfg TraceConfig) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return func(context.Context) error { return nil }, nil
	}
	if cfg.SampleRatio <= 0 || cfg.SampleRatio > 1 {
		cfg.SampleRatio = 1
	}
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.Endpoint))
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}
	res, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("deployment.environment.name", cfg.Environment),
	))
	if err != nil {
		return nil, fmt.Errorf("create trace resource: %w", err)
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
	)
	otel.SetTracerProvider(provider)
	return provider.Shutdown, nil
}

type PGXTracer struct{}

func NewPGXTracer() PGXTracer { return PGXTracer{} }

func (PGXTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	operation := "QUERY"
	if fields := strings.Fields(data.SQL); len(fields) > 0 {
		operation = strings.ToUpper(fields[0])
	}
	ctx, span := otel.Tracer(instrumentationName).Start(ctx, "postgres."+strings.ToLower(operation),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.system.name", "postgresql"), attribute.String("db.operation.name", operation)))
	return context.WithValue(ctx, pgxSpanKey{}, span)
}

func (PGXTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span, ok := ctx.Value(pgxSpanKey{}).(trace.Span)
	if !ok {
		return
	}
	defer span.End()
	if data.Err != nil && !errors.Is(data.Err, pgx.ErrNoRows) {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
	}
	if data.CommandTag.String() != "" {
		span.SetAttributes(attribute.String("db.response.status_code", data.CommandTag.String()))
	}
}

type pgxSpanKey struct{}
