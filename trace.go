package cache

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/writdev-alt/pkg-cache"

func startSpan(ctx context.Context, operation, key string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	base := []attribute.KeyValue{
		attribute.String("cache.key", key),
		attribute.String("cache.operation", operation),
	}
	base = append(base, attrs...)
	return otel.Tracer(tracerName).Start(ctx, "redis."+operation, trace.WithAttributes(base...))
}

func setHit(span trace.Span, hit bool) {
	span.SetAttributes(attribute.Bool("cache.hit", hit))
}

func setSource(span trace.Span, source string) {
	span.SetAttributes(attribute.String("cache.source", source))
}

func setTTL(span trace.Span, ttlSec float64) {
	span.SetAttributes(attribute.Float64("cache.ttl", ttlSec))
}

func recordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func ttlAttr(seconds float64) attribute.KeyValue {
	return attribute.String("cache.ttl", strconv.FormatFloat(seconds, 'f', -1, 64))
}
