package event

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
)

type OutboxStorageInterface interface {
	CreateEvent(ctx context.Context, event *entity.Event) error
	BatchCreateEvents(ctx context.Context, events entity.Events) error
}

type MetricsInterface interface {
	IncUnprocessedEventsCounter(count int)
}
