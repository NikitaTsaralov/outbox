package dto

import (
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	jsonUtils "github.com/NikitaTsaralov/transactional-outbox/pkg/jsonutils"
	"github.com/jackc/pgtype"
)

type Batch struct {
	EntityIDs       pgtype.TextArray
	IdempotencyKeys pgtype.TextArray
	Payloads        pgtype.TextArray
	TraceCarriers   pgtype.TextArray
	Topics          pgtype.TextArray
	Keys            pgtype.TextArray
	CreatedAts      pgtype.TimestampArray
	ExpiresAts      pgtype.TimestampArray
}

type transposedEventArray struct {
	EntityID       []string
	IdempotencyKey []string
	Payload        []string
	TraceCarrier   []string
	Topic          []string
	Key            []string
	CreatedAt      []time.Time
	ExpiresAt      []time.Time
}

func NewEventBatchFromEvents(e models.Events) Batch {
	transposed := transposedEventArray{
		EntityID:       make([]string, len(e)),
		IdempotencyKey: make([]string, len(e)),
		Payload:        make([]string, len(e)),
		TraceCarrier:   make([]string, len(e)),
		Topic:          make([]string, len(e)),
		Key:            make([]string, len(e)),
		CreatedAt:      make([]time.Time, len(e)),
		ExpiresAt:      make([]time.Time, len(e)),
	}

	for i, event := range e {
		transposed.EntityID[i] = event.EntityID
		transposed.IdempotencyKey[i] = event.IdempotencyKey
		transposed.Payload[i] = string(event.Payload)
		transposed.TraceCarrier[i] = string(jsonUtils.UnsafeMarshall(NewTraceCarrierFromContext(event.Context)))
		transposed.Topic[i] = event.Topic
		transposed.Key[i] = event.Key
		transposed.CreatedAt[i] = event.CreatedAt
		transposed.ExpiresAt[i] = event.ExpiresAt
	}

	var batch Batch

	batch.EntityIDs.Set(transposed.EntityID)             //nolint: errcheck // only error here can be because of type
	batch.IdempotencyKeys.Set(transposed.IdempotencyKey) //nolint: errcheck // only error here can be because of type
	batch.Payloads.Set(transposed.Payload)               //nolint: errcheck // only error here can be because of type
	batch.TraceCarriers.Set(transposed.TraceCarrier)     //nolint: errcheck // only error here can be because of type
	batch.Topics.Set(transposed.Topic)                   //nolint: errcheck // only error here can be because of type
	batch.Keys.Set(transposed.Key)                       //nolint: errcheck // only error here can be because of type
	batch.CreatedAts.Set(transposed.CreatedAt)           //nolint: errcheck // only error here can be because of type
	batch.ExpiresAts.Set(transposed.ExpiresAt)           //nolint: errcheck // only error here can be because of type

	return batch
}
