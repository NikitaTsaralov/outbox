package update

import (
	"context"

	"layout-example/internal/command/entity_name"
	"layout-example/internal/domain/entity"

	"github.com/NikitaTsaralov/utils/utils/trace"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"go.opentelemetry.io/otel"
)

type CommandHandler struct {
	storage   entity_name.StorageInterface
	txManager *manager.Manager
}

func NewCommandHandler(
	storage entity_name.StorageInterface,
	txManager *manager.Manager,
) *CommandHandler {
	return &CommandHandler{
		storage:   storage,
		txManager: txManager,
	}
}

func (h *CommandHandler) Execute(ctx context.Context, command *Command) error {
	ctx, span := otel.Tracer("").Start(ctx, "Command.Event.Create.Execute")
	defer span.End()

	domain := entity.NewEntity(
		entity.WithID(command.id),
		entity.WithValue(command.value),
	)

	err := h.txManager.Do(ctx, func(ctx context.Context) error {
		err := h.storage.Update(ctx, domain)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return trace.Wrapf(span, err, "Command.Event.Create.Execute.txManager.Do(command: %#v)", err)
	}

	return nil
}