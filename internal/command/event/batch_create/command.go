package batch_create

import (
	"encoding/json"
)

type EventPayload struct {
	IdempotencyKey string
	Payload        json.RawMessage
}

type Command struct {
	Events []*EventPayload
}

func NewCommand(command *Command) *Command {
	return &Command{Events: command.Events}
}
