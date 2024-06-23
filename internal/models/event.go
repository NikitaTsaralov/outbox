package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/guregu/null"
	"github.com/samber/lo"
)

type Event struct {
	ID             int64
	EntityID       string // preserve order and search purposes
	IdempotencyKey string // unique
	Topic          string
	Payload        json.RawMessage
	Context        context.Context
	CreatedAt      null.Time
	SentAt         null.Time
	TTL            time.Duration
}

type Events []Event // as it is entity we use pointers

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
