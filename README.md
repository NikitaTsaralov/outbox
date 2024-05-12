# github.com/NikitaTsaralov/transactional-outbox

![cqrs](./assets/ReliablePublication.png)

### Overview

> Transactional outbox is a pattern that ensures that messages are delivered to a message broker in a reliable way.

This project is my own implementation of transaction-outbox pattern for personal use.

### How to

1. Run ./examples/docker-compose.yml
2. Create topic in kafka

```shell
docker exec -it kafka kafka-topics --create --topic transactional-outbox --partitions 1 --replication-factor 1 --bootstrap-server localhost:29092
```

3. Run example (full example in ./examples/main.go)

```go
// init
client := transactionalOutbox.NewClient(
    postgresDB,
    kafkaClient,
    10000,
    manager.Must(trmsqlx.NewDefaultFactory(postgresDB)),
    trmsqlx.DefaultCtxGetter,
)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// create event
_, err := client.CreateEvent(ctx, entity.CreateEventCommand{
    EntityID:       uuid.New(),
    IdempotencyKey: "test",
    Payload:        []byte(`{"test": "test"}`),
    Topic:          "test",
})
if err != nil {
    logger.Fatalf("failed to create event: %v", err)
}

// batch create events
_, err := client.BatchCreateEvents(ctx, entity.BatchCreateEventCommand{
    {
        EntityID:       uuid.New(),
        IdempotencyKey: "test2",
        Payload:        []byte(`{"test": "test"}`),
        Topic:          "test",
    },
    {
        EntityID:       uuid.New(),
        IdempotencyKey: "test3",
        Payload:        []byte(`{"test": "test"}`),
        Topic:          "test",
    },
})
if err != nil {
    logger.Fatalf("failed to batch create events: %v", err)
}

// run message relay
client.RunMessageRelay(ctx)
```

### Linter

```shell
golangci-lint run  -v --config .golangci.yml
```