package create

import (
	"encoding/json"
)

type Command struct {
	IdempotencyKey string
	Payload        json.RawMessage
}

func NewCommand(
	IdempotencyKey string,
	Payload json.RawMessage,
) *Command {
	return &Command{
		IdempotencyKey: IdempotencyKey,
		Payload:        Payload,
	}
}
