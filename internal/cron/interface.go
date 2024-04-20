package cron

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/valueobject"
)

type OutboxStorageInterface interface {
	FetchUnprocessedEvents(ctx context.Context) (entity.Events, error)
	MarkEventsAsProcessed(ctx context.Context, ids valueobject.IDs) error
}

type OutboxPublisherInterface interface {
	PublishEvents(ctx context.Context, events entity.Events) error
}
