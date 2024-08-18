package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type CreateEventCommand struct {
	EntityID       string          // broker record key
	IdempotencyKey string          //
	Payload        json.RawMessage // broker record value
	Topic          string
	Key            string          // for uniform bytes partitioner
	Context        context.Context // individual context (4 batch especially)
	TTL            time.Duration   // time for event to stay in db
	CreatedAt      time.Time       // for delayed events or anything u need
}

func (o Outbox) CreateEventCommandHandler(ctx context.Context, command CreateEventCommand) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.CreateEventCommandHandler")
	defer span.End()

	if command.CreatedAt.IsZero() {
		command.CreatedAt = time.Now()
	}

	if command.Context == nil {
		command.Context = ctx
	}

	if command.Key == "" {
		command.Key = uuid.NewString() // disable uniform bytes partitioner
	}

	event := models.Event{
		EntityID:       command.EntityID,
		IdempotencyKey: command.IdempotencyKey,
		Topic:          command.Topic,
		Key:            command.Key,
		Payload:        command.Payload,
		Context:        command.Context,
		CreatedAt:      command.CreatedAt,
		ExpiresAt:      command.CreatedAt.Add(command.TTL),
	}

	return o.storage.CreateOutboxEvent(ctx, event)
}
