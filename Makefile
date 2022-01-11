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
	@docker volume rm timebench_data
	@rm -fv bin/*

# Database commands
migrate:
	@sleep 1
	@PGPASSWORD=password psql -U postgres -h localhost < db/migrations/cpu_usage.sql
	@PGPASSWORD=password psql -U postgres -h localhost -d homework -c "\COPY cpu_usage FROM db/migrations/cpu_usage.csv CSV HEADER"

database:
	@PGPASSWORD=password psql -U postgres -h localhost -d homework