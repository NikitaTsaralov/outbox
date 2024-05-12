package entity

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type CreateEventCommand struct {
	EntityID       uuid.UUID       `validate:"required"`
	IdempotencyKey string          `validate:"required"`
	Payload        json.RawMessage `validate:"required"`
	Topic          string          `validate:"required"`
}

type BatchCreateEventCommand []CreateEventCommand

func (e BatchCreateEventCommand) EntityIDs() []uuid.UUID {
	return lo.Map(e, func(event CreateEventCommand, _ int) uuid.UUID {
		return event.EntityID
	})
}
