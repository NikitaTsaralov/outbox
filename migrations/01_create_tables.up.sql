create extension if not exists "uuid-ossp";

create table if not exists outbox
(
    id              bigserial primary key,
    entity_id       text      not null,                                  -- for incident research
    idempotency_key text      not null unique default gen_random_uuid(), -- your own idempotency key for duplicate check
    topic           text      not null,
    key             text      not null        default gen_random_uuid(), -- uniform-bytes partitioner key
    payload         jsonb     not null,
    trace_carrier   jsonb     not null,
    created_at      timestamp not null        default now(),
    expires_at      timestamp not null        default now() + interval '1 day',
    sent_at         timestamp
);

create index outbox_entity_id_idx on outbox (entity_id); -- to search by entity_id
create index outbox_sent_at_idx on outbox (sent_at);