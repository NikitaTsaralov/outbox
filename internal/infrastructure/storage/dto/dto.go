package dto

import (
	"encoding/json"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
	"github.com/samber/lo"
)

type Event struct {
	ID             int64           `db:"id"`
	IdempotencyKey string          `db:"idempotency_key"`
	Payload        json.RawMessage `db:"payload"`
	TraceID        string          `db:"trace_id"`
	TraceCarrier   json.RawMessage `db:"trace_carrier"`
	Processed      bool            `db:"processed"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
}

func NewEventFromDomain(event *entity.Event) *Event {
	return &Event{
		ID:             event.ID,
		IdempotencyKey: event.IdempotencyKey,
		Payload:        event.Payload,
		TraceID:        event.TraceID,
		TraceCarrier:   event.TraceCarrier,
		Processed:      event.Processed,
		CreatedAt:      event.CreatedAt,
		UpdatedAt:      event.UpdatedAt,
	}
}

func (e *Event) ToDomain() *entity.Event {
	return &entity.Event{
		ID:             e.ID,
		IdempotencyKey: e.IdempotencyKey,
		Payload:        e.Payload,
		TraceID:        e.TraceID,
		TraceCarrier:   e.TraceCarrier,
		Processed:      e.Processed,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

type Events []*Event

func NewEventsFromDomain(events entity.Events) Events {
	return lo.Map(events, func(item *entity.Event, _ int) *Event {
		return NewEventFromDomain(item)
	})
}

func (e Events) ToDomain() entity.Events {
	return lo.Map(e, func(item *Event, _ int) *entity.Event {
		return item.ToDomain()
	})
}

type EventBatch struct {
	ID             []int64
	IdempotencyKey []string
	Payload        []string
	TraceID        []string
	TraceCarrier   []string
	Processed      []bool
	CreatedAt      []time.Time
	UpdatedAt      []time.Time
}

func NewEventBatchFromDomain(events entity.Events) *EventBatch {
	eventBatch := &EventBatch{
		ID:             make([]int64, len(events)),
		IdempotencyKey: make([]string, len(events)),
		Payload:        make([]string, len(events)),
		TraceID:        make([]string, len(events)),
		TraceCarrier:   make([]string, len(events)),
		Processed:      make([]bool, len(events)),
		CreatedAt:      make([]time.Time, len(events)),
		UpdatedAt:      make([]time.Time, len(events)),
	}

	for i, event := range events {
		eventBatch.ID[i] = event.ID
		eventBatch.IdempotencyKey[i] = event.IdempotencyKey
		eventBatch.Payload[i] = string(event.Payload)
		eventBatch.TraceID[i] = event.TraceID
		eventBatch.TraceCarrier[i] = string(event.TraceCarrier)
		eventBatch.Processed[i] = event.Processed
		eventBatch.CreatedAt[i] = event.CreatedAt
		eventBatch.UpdatedAt[i] = event.UpdatedAt
	}

	return eventBatch
}
