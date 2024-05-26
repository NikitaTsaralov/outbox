package postgres

const (
	queryCreateEvent = `
		insert into outbox (entity_id, idempotency_key, topic, payload, trace_carrier)
		values ($1, $2, $3, $4, $5)
		returning id;`

	queryBatchCreateEvent = `
		insert into outbox (entity_id, idempotency_key, topic, payload, trace_carrier)
		values (unnest($1::uuid[]),
				unnest($2::text[]),
				unnest($3::text[]),
				unnest($4::jsonb[]),
		        unnest($5::jsonb[]))
		returning id;`

	queryFetchUnprocessedEvents = `
		select id
		     , entity_id
		     , idempotency_key
		     , topic, payload
		     , trace_carrier
		     , created_at
		     , sent_at
			 , pg_try_advisory_lock('outbox'::regclass::oid::int, hashtext(entity_id)) as available
		from outbox 
		limit $1
		for update skip locked;`

	queryMarkEventsAsProcessed = `
		update outbox
		set sent_at = now()
		where id = any ($1);`

	queryDeleteProcessedEvents = `
		delete from outbox
			where sent_at < now() - $1 * interval '1 millisecond';`
)
