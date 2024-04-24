package create

import "github.com/google/uuid"

type Command struct {
	id    uuid.UUID
	value string
}

func NewCommand(id uuid.UUID, value string) *Command {
	return &Command{
		id:    id,
		value: value,
	}
}
