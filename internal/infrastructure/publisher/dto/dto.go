package dto

import (
	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
)

func NewKGORecordFromDomain(events entity.Events) []*kgo.Record {
	return lo.Map(events, func(event *entity.Event, _ int) *kgo.Record {
		return &kgo.Record{
			Key:   []byte(event.IdempotencyKey),
			Value: event.Payload,
			Headers: []kgo.RecordHeader{
				{
					Key:   "trace_carrier",
					Value: event.TraceCarrier,
				},
			},
		}
	})
}
