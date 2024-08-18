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
	Key            string
	Payload        json.RawMessage
	Context        context.Context
	CreatedAt      time.Time
	ExpiresAt      time.Time // TTL
	SentAt         null.Time
}

type Events []Event // as it is models we use pointers

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
