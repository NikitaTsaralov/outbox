CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table if not exists outbox
(
    id        bigserial primary key,
    entity_id uuid not null, -- can be int if you prefer
    idempotency_key text not null unique, -- your own idempotency key
    topic     text not null,
    payload jsonb not null,
    created_at timestamp not null default now(),
    sent_at timestamp
);

create index outbox_entity_id_idx on outbox(entity_id); -- to search by entity_id