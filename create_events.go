package outbox

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
)

func (o Client) CreateEvent(ctx context.Context, command CreateEventCommand) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.CreateEvent")
	defer span.End()

	return o.storage.CreateOutboxEvent(ctx, models.Event{
		EntityID:       command.EntityID,
		IdempotencyKey: command.IdempotencyKey,
		Topic:          command.Topic,
		Payload:        command.Payload,
		Context:        ctx,
		TTL:            command.TTL,
	})
}

func (o Client) BatchCreateEvents(ctx context.Context, command BatchCreateEventCommand) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.BatchCreateEvents")
	defer span.End()

	return o.storage.BatchCreateOutboxEvents(ctx, lo.Map(command, func(item CreateEventCommand, _ int) models.Event {
		return models.Event{
			EntityID:       item.EntityID,
			IdempotencyKey: item.IdempotencyKey,
			Topic:          item.Topic,
			Payload:        item.Payload,
			Context:        ctx,
			TTL:            item.TTL,
		}
	}))
}
