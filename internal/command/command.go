package command

import (
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/batch_create"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/create"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
)

type EventCommand struct {
	Create      *create.CommandHandler
	BatchCreate *batch_create.CommandHandler
}

func NewOutboxCommand(
	storage event.OutboxStorageInterface,
	txManager *manager.Manager,
) *EventCommand {
	return &EventCommand{
		Create:      create.NewCommandHandler(storage, txManager),
		BatchCreate: batch_create.NewCommandHandler(storage, txManager),
	}
}
