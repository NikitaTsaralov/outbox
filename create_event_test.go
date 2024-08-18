package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/repository/postgres/dto"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (s *OutboxTestSuite) Test_CreateEvent() {
	tests := []struct {
		name string
		data CreateEventCommand
	}{
		{
			name: "test 1",
			data: CreateEventCommand{
				EntityID:       uuid.NewString(),
				IdempotencyKey: uuid.NewString(),
				Payload:        json.RawMessage(`{"1": "2"}`),
				Topic:          "transactional-outbox",
				Key:            uuid.NewString(),
				TTL:            time.Second * 10,
				CreatedAt:      time.Now().UTC(),
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			id, err := s.outbox.CreateEvent(context.Background(), test.data)
			s.Require().Nil(err)
			s.Require().NotNil(id)

			var createdEventDto dto.Events

			err = s.db.Select(&createdEventDto, `
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
				order by entity_id`, pq.Int64Array([]int64{id}))
			s.Require().Nil(err)
			s.Require().Equal(len(createdEventDto), 1)

			createdEvent := createdEventDto.ToModel()

			s.Require().NotNil(createdEvent[0].ID)
			s.Require().Equal(createdEvent[0].ID, id)
			s.Require().Equal(createdEvent[0].EntityID, test.data.EntityID)
			s.Require().Equal(createdEvent[0].IdempotencyKey, test.data.IdempotencyKey)
			s.Require().Equal(createdEvent[0].Payload, test.data.Payload)
			s.Require().Equal(createdEvent[0].Topic, test.data.Topic)
			s.Require().Equal(createdEvent[0].Key, test.data.Key)
			s.Require().NotNil(createdEvent[0].Context)
			s.Require().NotNil(createdEvent[0].CreatedAt)
			s.Require().Equal(createdEvent[0].CreatedAt, test.data.CreatedAt)
			s.Require().Equal(createdEvent[0].ExpiresAt, test.data.CreatedAt.Add(test.data.TTL))
			s.Require().False(createdEvent[0].SentAt.Valid)
		})
	}
}
