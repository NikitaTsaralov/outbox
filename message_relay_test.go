package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/repository/postgres/dto"
)

func (s *OutboxTestSuite) Test_RunMessageRelay() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.outbox.RunMessageRelay(ctx)

	s.Test_BatchCreateEvent() // create some event (it's bad cause tests are not atomic)

	time.Sleep(5 * time.Second) // this time is enough

	var dtoCreatedEvents dto.Events

	err := s.db.Select(&dtoCreatedEvents, `
		select id
		     , entity_id
		     , idempotency_key
		     , topic, payload
		     , trace_carrier
		     , created_at
		     , sent_at
			 , expires_at
		from outbox 
		order by entity_id`)
	s.Require().Nil(err)

	createdEvents := dtoCreatedEvents.ToModel()

	for _, createdEvent := range createdEvents {
		s.Require().True(createdEvent.SentAt.Valid)
		s.Require().NotNil(createdEvent.SentAt)
	}
}

func (s *OutboxTestSuite) Test_messageRelayTick() {
	s.Test_BatchCreateEvent() // create some event (it's bad cause tests are not atomic)
	s.outbox.messageRelayTick()

	var dtoCreatedEvents dto.Events

	err := s.db.Select(&dtoCreatedEvents, `
		select id
		     , entity_id
		     , idempotency_key
		     , topic, payload
		     , trace_carrier
		     , created_at
		     , sent_at
			 , expires_at
		from outbox 
		order by entity_id`)
	s.Require().Nil(err)

	createdEvents := dtoCreatedEvents.ToModel()

	for _, createdEvent := range createdEvents {
		s.Require().True(createdEvent.SentAt.Valid)
		s.Require().NotNil(createdEvent.SentAt)
	}
}
