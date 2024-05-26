package tests

import (
	"context"

	outbox "github.com/NikitaTsaralov/transactional-outbox"
)

// TransactionalOutbox interface describes public functions u obtain using this package
type TransactionalOutbox interface {
	CreateEvent(ctx context.Context, event outbox.CreateEventCommand) (int64, error)
	BatchCreateEvents(ctx context.Context, events outbox.BatchCreateEventCommand) ([]int64, error)
	RunMessageRelay(ctx context.Context)
	RunGarbageCollector(ctx context.Context)
}
