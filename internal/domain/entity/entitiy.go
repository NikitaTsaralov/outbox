package entity

import (
	"encoding/json"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/valueobject"
	"github.com/samber/lo"
)

type Event struct {
	ID             int64
	IdempotencyKey string
	Payload        json.RawMessage
	TraceID        string
	TraceCarrier   json.RawMessage
	Processed      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Events []*Event

func (e Events) IDs() valueobject.IDs {
	return lo.Map(e, func(item *Event, _ int) valueobject.ID {
		return valueobject.ID(item.ID)
	})
}
