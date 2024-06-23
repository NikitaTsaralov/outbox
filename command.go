package outbox

import (
	"context"
	"encoding/json"
	"time"
)

type CreateEventCommand struct {
	EntityID       string          // broker record key
	IdempotencyKey string          //
	Payload        json.RawMessage // broker record value
	Topic          string
	Context        context.Context // individual context (4 batch especially)
	TTL            time.Duration   // time to live
}

type BatchCreateEventCommand []CreateEventCommand
