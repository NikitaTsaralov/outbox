package dto

import (
	"encoding/json"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/guregu/null"
	"github.com/samber/lo"
)

type Event struct {
	ID             int64           `db:"id"`
	EntityID       string          `db:"entity_id"`
	IdempotencyKey string          `db:"idempotency_key"`
	Topic          string          `db:"topic"`
	Payload        json.RawMessage `db:"payload"`
	TraceCarrier   TraceCarrier    `db:"trace_carrier"`
	CreatedAt      null.Time       `db:"created_at"`
	SentAt         null.Time       `db:"sent_at"`
	TTL            int64           `db:"ttl"`
	Available      bool            `db:"available"`
}

func NewEventFromModel(e models.Event) Event {
	return Event{
		ID:             e.ID,
		EntityID:       e.EntityID,
		IdempotencyKey: e.IdempotencyKey,
		Topic:          e.Topic,
		Payload:        e.Payload,
		TraceCarrier:   NewTraceCarrierFromContext(e.Context),
		CreatedAt:      e.CreatedAt,
		SentAt:         e.SentAt,
		TTL:            e.TTL.Milliseconds(),
	}
}

func (e Event) ToModel() models.Event {
	return models.Event{
		ID:             e.ID,
		EntityID:       e.EntityID,
		IdempotencyKey: e.IdempotencyKey,
		Topic:          e.Topic,
		Payload:        e.Payload,
		Context:        e.TraceCarrier.Context(),
		CreatedAt:      e.CreatedAt,
		SentAt:         e.SentAt,
		TTL:            time.Duration(e.TTL) * time.Millisecond,
	}
}

type Events []Event

func (e Events) ToModel() models.Events {
	return lo.Map(e, func(item Event, _ int) models.Event {
		return item.ToModel()
	})
}
