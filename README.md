# transactional-outbox

### Overview

> Pattern explanation: https://microservices.io/patterns/data/transactional-outbox.html

![transactional_outbox](./src/ReliablePublication.png)

### Idea Explanation

This package includes all outbox interaction
* creating outbox message (single record & batch)
* publishing outbox to kafka via message relay worker

### Example

```go
cfg := &config.MessageRelay{
    Enabled:   true,
    Timeout:   1000, // 1 s
    BatchSize: 10,
}

// start message relay worker
s := application.NewService(cfg, nil, nil, nil)
s.Start()

// graceful shutdown
s.Stop()
```

### Project layout explanation. We use:

* hexagonal architecture 
* DDD features (domain models, value objects, DTOs)
* CQRS pattern (commands)

### TODOs:

* use `github.com/jackc/pgx/v5/pgxpool` instead of `github.com/jmoiron/sqlx`

### Acknowledgments:

* https://microservices.io
* https://github.com/twmb/franz-go
* https://github.com/ozontech/file.d