package outbox

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/repository/postgres/dto"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (s *OutboxTestSuite) Test_BatchCreateEvent() {
	tests := []struct {
		name string
		data BatchCreateEventsCommand
	}{
		{
			name: "test 1",
			data: BatchCreateEventsCommand{
				CreateEventCommand{
					EntityID:       uuid.NewString(),
					IdempotencyKey: uuid.NewString(),
					Payload:        json.RawMessage(`{"3": "4"}`),
					Topic:          "transactional-outbox",
					Key:            uuid.NewString(),
					TTL:            10 * time.Second,
					CreatedAt:      time.Now().UTC(),
				},
				CreateEventCommand{
					EntityID:       uuid.NewString(),
					IdempotencyKey: uuid.NewString(),
					Payload:        json.RawMessage(`{"5": "6"}`),
					Topic:          "transactional-outbox",
					Key:            uuid.NewString(),
					TTL:            10 * time.Second,
					CreatedAt:      time.Now().UTC(),
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			sort.Slice(test.data, func(i, j int) bool {
				return test.data[i].EntityID < test.data[j].EntityID
			})

			ids, err := s.outbox.BatchCreateEvents(context.Background(), test.data)
			s.Require().Nil(err)
			s.Require().NotNil(ids)

			var createdEventsDto dto.Events

			err = s.db.Select(&createdEventsDto, `
				select id
					 , entity_id
					 , idempotency_key
					 , topic
					 , key
					 , payload
					 , trace_carrier
					 , created_at
					 , sent_at
					 , expires_at
				from outbox 
				where id = any($1)
				order by entity_id`, pq.Int64Array(ids))
			s.Require().Nil(err)
			s.Require().Equal(len(ids), len(test.data))

			createdEvents := createdEventsDto.ToModel()

			for i := 0; i < len(ids); i++ {
				s.Require().NotNil(createdEvents[i].ID)
				s.Require().Equal(createdEvents[i].ID, ids[i])
				s.Require().Equal(createdEvents[i].EntityID, test.data[i].EntityID)
				s.Require().Equal(createdEvents[i].IdempotencyKey, test.data[i].IdempotencyKey)
				s.Require().Equal(createdEvents[i].Payload, test.data[i].Payload)
				s.Require().Equal(createdEvents[i].Topic, test.data[i].Topic)
				s.Require().Equal(createdEvents[i].Key, test.data[i].Key)
				s.Require().NotNil(createdEvents[i].Context)
				s.Require().Equal(createdEvents[i].CreatedAt, test.data[i].CreatedAt)
				s.Require().Equal(createdEvents[i].ExpiresAt, test.data[i].CreatedAt.Add(test.data[i].TTL))
				s.Require().False(createdEvents[i].SentAt.Valid)
			}
		})
	}
}
