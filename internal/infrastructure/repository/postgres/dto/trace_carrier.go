package dto

import (
	"context"
	"encoding/json"

	"github.com/NikitaTsaralov/outbox/internal/infrastructure/repository/postgres/errors"
	"go.opentelemetry.io/otel"
)

// TraceCarrier is a TextMapCarrier that uses a map held in memory as a repository
// medium for propagated key-value pairs.
type TraceCarrier map[string]string

func NewTraceCarrierFromContext(ctx context.Context) TraceCarrier {
	carrier := make(TraceCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier
}

// Get returns the value associated with the passed key.
func (t TraceCarrier) Get(key string) string {
	return t[key]
}

// Set stores the key-value pair.
func (t TraceCarrier) Set(key, value string) {
	t[key] = value
}

// Keys lists the keys stored in this carrier.
func (t TraceCarrier) Keys() []string {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}

	return keys
}

func (t *TraceCarrier) Context() context.Context {
	ctx := context.Background()
	otel.GetTextMapPropagator().Extract(ctx, t)

	return ctx
}

func (t *TraceCarrier) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return errors.ErrTypeAssertionFailed
	}

	return json.Unmarshal(b, &t)
}
