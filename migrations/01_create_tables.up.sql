create table if not exists outbox
(
    id              bigserial primary key,
    entity_id       text      not null,        -- can be int if you prefer
    idempotency_key text      not null unique, -- your own idempotency key
    topic           text      not null,
    payload         jsonb     not null,
    trace_carrier   jsonb     not null,
    created_at      timestamp not null default now(),
    sent_at         timestamp
);

create index outbox_entity_id_idx on outbox (entity_id); -- to search by entity_id
create index sent_at_idx on outbox (sent_at); -- to collect garbage and order by