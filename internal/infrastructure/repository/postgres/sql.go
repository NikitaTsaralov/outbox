package postgres

const (
	queryCreateEvent = `
		insert into outbox (entity_id, idempotency_key, topic, key, payload, trace_carrier, created_at, expires_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id;`

	queryBatchCreateEvent = `
		insert into outbox (entity_id, idempotency_key, topic, key, payload, trace_carrier, created_at, expires_at)
		values (unnest($1::uuid[]),
				unnest($2::text[]),
				unnest($3::text[]),
				unnest($4::text[]),
				unnest($5::jsonb[]),
		        unnest($6::jsonb[]),
		        unnest($7::timestamptz[]),
		        unnest($8::timestamptz[]))
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

	queryCountExpiredEvents = `
		select count(*)
		from outbox
		where sent_at is not null and expires_at < now();`

	queryFetchExpiredEvents = `
		select id
			 , entity_id
			 , idempotency_key
			 , topic
			 , payload
			 , trace_carrier
			 , created_at
			 , sent_at
			 , key
		from outbox
		where sent_at is not null and expires_at < now()
		order by id
		limit $1 for update skip locked;`

	queryDeleteEventsByIDs = `
		delete from outbox
		where id = any($1);`
)
