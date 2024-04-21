package postgres

const (
	queryCreateEvent = `
		insert
		  into outbox
			(idempotency_key, payload, trace_id, trace_carrier, processed, created_at, updated_at)
		  values
			($1, $2, $3, $4, $5, $6, $7);`

	queryBatchCreateEvent = `
		insert
		  into outbox
			(idempotency_key, payload, trace_id, trace_carrier, processed)
		select *
		  from unnest(
				   $1::text[],
				   $2::jsonb[],
				   $3::text[],
				   $4::jsonb[],
				   $5::bool[]
			   ) as t (
			idempotency_key, payload, trace_id, trace_carrier, processed
			);`

	queryFetchUnprocessedEvents = `
		select id, payload, created_at from outbox
		where processed = false
		order by created_at
		limit $1 for update skip locked;`

	queryMarkEventsAsProcessed = `
		update outbox set processed = true
		where outbox.id = any($1);`
)
