.PHONY: build run migrate database down

# CLI commands
build:
	@GOFLAGS=-mod=mod go build -o bin/timebench main.go

run:
	@GOFLAGS=-mod=mod go run main.go

# Docker commands
up:
	@docker-compose up -d

down:
	@docker-compose down

clean:
	@docker-compose down
	@docker volume prune -f

# Database commands
migrate:
	@PGPASSWORD=password psql -U postgres -h localhost < db/migrations/cpu_usage.sql
	@PGPASSWORD=password psql -U postgres -h localhost -d homework -c "\COPY cpu_usage FROM db/migrations/cpu_usage.csv CSV HEADER"
	@PGPASSWORD=password psql -U postgres -h localhost -d homework -c "SELECT pg_stat_statements_reset();"

database:
	@PGPASSWORD=password psql -U postgres -h localhost -d homework