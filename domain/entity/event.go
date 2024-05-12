package entity

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/lib/pq"
	"github.com/samber/lo"
)

type Event struct {
	ID             int64           `db:"id"`
	EntityID       uuid.UUID       `db:"entity_id"`
	IdempotencyKey string          `db:"idempotency_key"`
	Topic          string          `db:"topic"`
	Payload        json.RawMessage `db:"payload"`
	CreatedAt      null.Time       `db:"created_at"`
	SentAt         null.Time       `db:"sent_at"`
}

type EventBatch struct {
	EntityIDs       pq.StringArray
	IdempotencyKeys pq.StringArray
	Payloads        pq.StringArray
	Topics          pq.StringArray
}

func NewEventBatchFromCommand(e BatchCreateEventCommand) EventBatch {
	batch := EventBatch{
		EntityIDs:       make(pq.StringArray, 0, len(e)),
		IdempotencyKeys: make(pq.StringArray, 0, len(e)),
		Payloads:        make(pq.StringArray, 0, len(e)),
		Topics:          make(pq.StringArray, 0, len(e)),
	}

	for i, event := range e {
		batch.EntityIDs[i] = event.EntityID.String()
		batch.IdempotencyKeys[i] = event.IdempotencyKey
		batch.Payloads[i] = string(event.Payload)
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
