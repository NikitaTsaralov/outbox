package postgres

const (
	queryCreate = `insert into some_table (name, ack_policy) values ($1, $2) returning id`

	queryUpdate = `update some_table set name = $1, ack_policy = $2 where id = $3`

	queryGetByID = `select * from some_table where id = $1`

	queryFetch = `select * from some_table`
)
