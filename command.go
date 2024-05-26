package outbox

import (
	"encoding/json"
)

type CreateEventCommand struct {
	EntityID       string
	IdempotencyKey string
	Payload        json.RawMessage
	Topic          string
}

type BatchCreateEventCommand []CreateEventCommand
