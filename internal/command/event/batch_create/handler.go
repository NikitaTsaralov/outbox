package batch_create

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event"
	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/valueobject"
	"github.com/NikitaTsaralov/utils/utils/trace"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
)

type CommandHandler struct {
	storage   event.OutboxStorageInterface
	metrics   event.MetricsInterface
	txManager *manager.Manager
}

func NewCommandHandler(
	storage event.OutboxStorageInterface,
	metrics event.MetricsInterface,
	txManager *manager.Manager,
) *CommandHandler {
	return &CommandHandler{
		storage:   storage,
		metrics:   metrics,
		txManager: txManager,
	}
}

func (h *CommandHandler) Execute(ctx context.Context, command *Command) error {
	ctx, span := otel.Tracer("").Start(ctx, "Command.Event.Create.Execute")
	defer span.End()

	events := lo.Map(command.Events, func(event *EventPayload, _ int) *entity.Event {
		return &entity.Event{
			IdempotencyKey: event.IdempotencyKey,
			Payload:        event.Payload,
			TraceID:        span.SpanContext().TraceID().String(),
			TraceCarrier:   valueobject.ExtractTraceCarrierToJSON(ctx),
			Processed:      false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	})

	err := h.txManager.Do(ctx, func(ctx context.Context) error {
		err := h.storage.BatchCreateEvents(ctx, events)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return trace.Wrapf(span, err, "Command.Event.Create.Execute(events: %v)", events)
	}

	h.metrics.IncUnprocessedEventsCounter(len(events))

	return nil
}
