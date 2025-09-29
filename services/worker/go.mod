module hafiztri123/worker-link-shortener

replace hpj/hv1-link-shortener/shared => ../../shared

go 1.23.0

toolchain go1.23.5

require (
	github.com/jackc/pgx/v5 v5.5.4
	github.com/joho/godotenv v1.5.1
	github.com/rabbitmq/amqp091-go v1.10.0
	hpj/hv1-link-shortener/shared v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)
