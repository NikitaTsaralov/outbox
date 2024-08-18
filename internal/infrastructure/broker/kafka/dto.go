package kafka

import (
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
)

const idempotencyKey = "idempotencyKey"

func newRecordFromModel(e models.Event) *kgo.Record {
	return &kgo.Record{
		Key:   []byte(e.Key), // preserve orders
		Value: e.Payload,
		Headers: []kgo.RecordHeader{
			{
				Key:   idempotencyKey,
				Value: []byte(e.IdempotencyKey),
			},
		},
		Timestamp:     time.Time{},
		Topic:         e.Topic,
		Partition:     0,
		Attrs:         kgo.RecordAttrs{},
		ProducerEpoch: 0,
		ProducerID:    0,
		LeaderEpoch:   0,
		Offset:        0,
		Context:       e.Context,
	}
}

func newRecordsFromModel(e models.Events) []*kgo.Record {
	return lo.Map(e, func(item models.Event, _ int) *kgo.Record {
		return newRecordFromModel(item)
	})
}
