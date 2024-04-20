package kafka

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/publisher/dto"
	"github.com/NikitaTsaralov/utils/utils/trace"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type Publisher struct {
	client *kgo.Client
}

func NewPublisher(client *kgo.Client) *Publisher {
	return &Publisher{
		client: client,
	}
}

func (p *Publisher) PublishEvents(ctx context.Context, events entity.Events) error {
	ctx, span := otel.Tracer("").Start(ctx, "Publisher.PublishEvents")
	defer span.End()

	dtoRecords := dto.NewKGORecordFromDomain(events)
	results := p.client.ProduceSync(ctx, dtoRecords...)

	err := results.FirstErr()
	if err != nil {
		return trace.Wrapf(span, err, "Publisher.PublishEvents.ProduceSync(events: %v)", events)
	}

	return nil
}
