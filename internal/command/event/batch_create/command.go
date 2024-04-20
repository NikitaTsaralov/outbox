package batch_create

import "github.com/NikitaTsaralov/transactional-outbox/internal/command/event/create"

type Command struct {
	events []*create.Command
}

func NewCommand(events []*create.Command) *Command {
	return &Command{events: events}
}
