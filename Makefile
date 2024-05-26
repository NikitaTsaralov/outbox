docker:
	docker-compose up -d

migrate_up:
	migrate -database postgres://root:dev@localhost:5432/postgres?sslmode=disable -path migrations up 1

migrate_down:
	migrate -database postgres://root:dev@localhost:5432/postgres?sslmode=disable -path migrations down 1

create_topic:
	docker exec -it transactional-outbox-kafka kafka-topics --create --topic transactional-outbox --partitions 1 --replication-factor 1 --bootstrap-server localhost:29092

test:
	go test ./..

linter:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run