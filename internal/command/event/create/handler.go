package create

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event"
	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/valueobject"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"go.opentelemetry.io/otel"
)

type CommandHandler struct {
	storage   event.OutboxStorageInterface
	txManager *manager.Manager
}

func NewCommandHandler(
	storage event.OutboxStorageInterface,
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

	err := h.storage.CreateEvent(ctx, &entity.Event{
		IdempotencyKey: command.IdempotencyKey,
		Payload:        command.Payload,
		TraceID:        span.SpanContext().TraceID().String(),
		TraceCarrier:   valueobject.ExtractTraceCarrierToJSON(ctx),
		Processed:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}
