package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
)

type BatchCreateEventsCommand []CreateEventCommand

func (o Outbox) BatchCreateEvents(
	ctx context.Context,
	command BatchCreateEventsCommand,
) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.BatchCreateEvents")
	defer span.End()

	events := lo.Map(command, func(item CreateEventCommand, _ int) models.Event {
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now()
		}

		if item.Context == nil {
			item.Context = ctx
		}

		if item.Key == "" {
			item.Key = uuid.NewString() // disable uniform bytes partitioner
		}

		return models.Event{
			EntityID:       item.EntityID,
			IdempotencyKey: item.IdempotencyKey,
			Topic:          item.Topic,
			Key:            item.Key,
			Payload:        item.Payload,
			Context:        item.Context,
			CreatedAt:      item.CreatedAt,
			ExpiresAt:      item.CreatedAt.Add(item.TTL),
		}
	})

	return o.storage.BatchCreateOutboxEvents(ctx, events)
}
