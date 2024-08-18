package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/repository/postgres/dto"
)

func (s *OutboxTestSuite) Test_RunGarbageCollector() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Test_BatchCreateEvent() // create some event (it's bad cause tests are not atomic)

	time.Sleep(11 * time.Second) // skip relay

	var dtoEventsBefore dto.Events

	err := s.db.Select(&dtoEventsBefore, `
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

	eventsBefore := dtoEventsBefore.ToModel()

	eventsToDeleteCount := 0

	for _, eventBefore := range eventsBefore {
		if eventBefore.SentAt.Valid && time.Since(eventBefore.SentAt.Time) > eventTTL {
			eventsToDeleteCount++
		}
	}

	go s.outbox.RunGarbageCollector(ctx)

	time.Sleep(5 * time.Second)

	var dtoCreatedEvents dto.Events

	err = s.db.Select(&dtoCreatedEvents, `
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
	s.Require().Equal(len(eventsBefore)-eventsToDeleteCount, len(dtoCreatedEvents))
}

func (s *OutboxTestSuite) Test_garbageCollectorTick() {
	s.Test_BatchCreateEvent() // create some event (it's bad cause tests are not atomic)
	_, err := s.db.Exec(`
			update outbox
			set sent_at = now(),
			    expires_at = now() - interval '1 hour'`) // send & expire
	s.Require().NoError(err)
	s.outbox.garbageCollectorTick() // delete

	var count int

	err = s.db.Get(&count, `select count(*) from outbox`)
	s.Require().NoError(err)
	s.Equal(count, 0)
}
