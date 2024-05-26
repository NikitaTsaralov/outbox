package interfaces

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/models/event"
)

type Broker interface {
	PublishEvents(ctx context.Context, events event.Events) ([]int64, error)
}
