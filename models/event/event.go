package event

import (
	"encoding/json"

	jsonUtils "github.com/NikitaTsaralov/transactional-outbox/pkg/jsonutils"
	traceCarrier "github.com/NikitaTsaralov/transactional-outbox/pkg/tracecarrier"
	"github.com/guregu/null"
	"github.com/lib/pq"
	"github.com/samber/lo"
)

type Event struct {
	ID             int64                     `db:"id"`
	EntityID       string                    `db:"entity_id"`
	IdempotencyKey string                    `db:"idempotency_key"`
	Topic          string                    `db:"topic"`
	Payload        json.RawMessage           `db:"payload"`
	TraceCarrier   traceCarrier.TraceCarrier `db:"trace_carrier"`
	CreatedAt      null.Time                 `db:"created_at"`
	SentAt         null.Time                 `db:"sent_at"`
	Available      bool                      `db:"available"`
}

type Batch struct {
	EntityIDs       pq.StringArray
	IdempotencyKeys pq.StringArray
	Payloads        pq.StringArray
	TraceCarriers   pq.StringArray
	Topics          pq.StringArray
}

func NewEventBatchFromEvents(e Events) Batch {
	batch := Batch{
		EntityIDs:       make(pq.StringArray, len(e)),
		IdempotencyKeys: make(pq.StringArray, len(e)),
		Payloads:        make(pq.StringArray, len(e)),
		TraceCarriers:   make(pq.StringArray, len(e)),
		Topics:          make(pq.StringArray, len(e)),
	}

	for i, event := range e {
		batch.EntityIDs[i] = event.EntityID
		batch.IdempotencyKeys[i] = event.IdempotencyKey
		batch.Payloads[i] = string(event.Payload)
		batch.TraceCarriers[i] = string(jsonUtils.UnsafeMarshall(event.TraceCarrier))
		batch.Topics[i] = event.Topic
	}

	return batch
}

type Events []Event

func (e Events) IDs() []int64 {
	return lo.Map(e, func(event Event, _ int) int64 {
		return event.ID
	})
}

func (e Events) EntityIDs() []string {
	return lo.Map(e, func(event Event, _ int) string {
		return event.EntityID
	})
}
