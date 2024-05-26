package tests

const (
	queryFetchEventsByIDs = `
		select id
		     , entity_id
		     , idempotency_key
		     , topic, payload
		     , trace_carrier
		     , created_at
		     , sent_at
		from outbox 
		where id = any($1)
		order by entity_id`

	queryFetchAll = `
		select id
		     , entity_id
		     , idempotency_key
		     , topic, payload
		     , trace_carrier
		     , created_at
		     , sent_at
		from outbox 
		order by entity_id`

	queryDeleteAll = `delete from outbox;`
)
