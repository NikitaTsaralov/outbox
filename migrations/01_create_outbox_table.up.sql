create table outbox
(
    id              bigserial primary key,
    idempotency_key text        not null unique, -- can b used to maintain event orders
    payload         jsonb       not null,
    trace_id        text,
    trace_carrier   jsonb       not null,
    processed       bool        not null default false,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now()
);

create index created_index on outbox (created_at);
create index idempotency_index on outbox (idempotency_key);