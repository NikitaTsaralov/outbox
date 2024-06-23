package dto

import (
	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	jsonUtils "github.com/NikitaTsaralov/transactional-outbox/pkg/jsonutils"
	"github.com/lib/pq"
)

type Batch struct {
	EntityIDs       pq.StringArray
	IdempotencyKeys pq.StringArray
	Payloads        pq.StringArray
	TraceCarriers   pq.StringArray
	Topics          pq.StringArray
	TTL             pq.Int64Array
}

func NewEventBatchFromEvents(e models.Events) Batch {
	batch := Batch{
		EntityIDs:       make(pq.StringArray, len(e)),
		IdempotencyKeys: make(pq.StringArray, len(e)),
		Payloads:        make(pq.StringArray, len(e)),
		TraceCarriers:   make(pq.StringArray, len(e)),
		Topics:          make(pq.StringArray, len(e)),
		TTL:             make(pq.Int64Array, len(e)),
	}

	for i, event := range e {
		batch.EntityIDs[i] = event.EntityID
		batch.IdempotencyKeys[i] = event.IdempotencyKey
		batch.Payloads[i] = string(event.Payload)
		batch.TraceCarriers[i] = string(jsonUtils.UnsafeMarshall(NewTraceCarrierFromContext(event.Context)))
		batch.Topics[i] = event.Topic
		batch.TTL[i] = event.TTL.Milliseconds()
	}

	return batch
}
