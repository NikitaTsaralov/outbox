# cqrs-layout-example

![cqrs](./assets/cqrs.png)

### Overview

> CQRS stands for Command and Query Responsibility Segregation, a pattern that separates read and update operations for a data store.

This project is example of cqrs pattern implementation on golang.
Can be used as project starter for crud based applications.
You just need to replace entity_name with your own domain entity

### Layout features

* Hexagonal architecture
* DTO (Data Transfer Object) for request and response
* DDD (Domain Driven Design) for domain entity
* CQRS (Command Query Responsibility Segregation) for separation of read and write operation (replication support)