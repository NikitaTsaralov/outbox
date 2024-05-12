package broker

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/domain/entity"
)

type Broker interface {
	PublishEvents(ctx context.Context, events entity.Events) ([]int64, error)
}
