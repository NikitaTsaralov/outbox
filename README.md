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
package main

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/application"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/batch_create"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/create"
	"github.com/google/uuid"
)

func main() {
	s := application.NewService(nil, nil, nil, nil)

	err := s.CreateEvent(context.Background(), &create.Command{
		IdempotencyKey: uuid.NewString(),
		Payload:        []byte("{\"a\": \"b\"}"),
	})
	if err != nil {
		return
	}

	err = s.BatchCreateEvent(context.Background(), &batch_create.Command{
		{
			IdempotencyKey: uuid.NewString(),
			Payload:        []byte("{\"c\": \"d\"}"),
		},
		{
			IdempotencyKey: uuid.NewString(),
			Payload:        []byte("{\"e\": \"f\"}"),
		},
	})
	if err != nil {
		return
	}

	s.Start()
	defer s.Stop()
}
```

### Project layout explanation. We use:

* hexagonal architecture 
* DDD features (domain models, value objects, DTOs)
* CQRS pattern (commands)

### TODOs:

* use `github.com/jackc/pgx/v5/pgxpool` instead of `github.com/jmoiron/sqlx`
* replace `pq.Array` with `pgtype.Array`

### Acknowledgments:

* https://microservices.io
* https://github.com/twmb/franz-go
* https://github.com/ozontech/file.d