package postgres

const (
	queryCreateEvent = `
		insert into outbox (entity_id, idempotency_key, topic, payload)
		values ($1, $2, $3, $4)
		returning id;`

	queryBatchCreateEvent = `
		insert into outbox (entity_id, idempotency_key, topic, payload)
		values (unnest($1::uuid[]),
				unnest($2::text[]),
				unnest($3::text[]),
				unnest($4::jsonb[]))
		returning id;`

	queryFetchUnprocessedEvents = `
		select id, entity_id, idempotency_key, topic, payload, created_at, sent_at
		from outbox;`

	queryMarkEventsAsProcessed = `
		update outbox
		set sent_at = now()
		where id = any ($1);`
)
