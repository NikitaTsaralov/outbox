package create

import (
	"encoding/json"
)

type Command struct {
	ID             int64
	IdempotencyKey string
	Payload        json.RawMessage
}

func NewCommand(
	ID int64,
	IdempotencyKey string,
	Payload json.RawMessage,
) *Command {
	return &Command{
		ID:             ID,
		IdempotencyKey: IdempotencyKey,
		Payload:        Payload,
	}
}
