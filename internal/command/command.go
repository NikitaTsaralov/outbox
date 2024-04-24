package command

import (
	"layout-example/internal/command/entity_name"
	"layout-example/internal/command/entity_name/create"
	"layout-example/internal/command/entity_name/update"

	"github.com/avito-tech/go-transaction-manager/trm/manager"
)

type EntityCommand struct {
	Create *create.CommandHandler
	Update *update.CommandHandler
}

func NewOutboxCommand(
	storage entity_name.StorageInterface,
	txManager *manager.Manager,
) *EntityCommand {
	return &EntityCommand{
		Create: create.NewCommandHandler(storage, txManager),
		Update: update.NewCommandHandler(storage, txManager),
	}
}
