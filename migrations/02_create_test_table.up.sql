create table test (
    id serial primary key,
    payload jsonb not null,
    err bool check ( not err ) -- to use in test for transaction errors
)