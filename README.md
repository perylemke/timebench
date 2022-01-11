# Timebench

A small CLI to generate statistics of TimescaleDB.

# Considerations

First, It's a very fun challenge, because Go it's not my principal language, but I can developing without problems.

In a first moment, I choose collect the metrics in a `pg_stat_statements`, but in talk with Sam It's better taking the metrics on application and is more easy too.

I presume to run this test it's necessary to install Go, Docker and Docker Compose, above links to install:

- [Go](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/engine/install/)
- [Docker Compose](https://docs.docker.com/compose/install/)

This CLI has some improvements, like refactor in a new functions and implement unit tests too.

And have a bug, where the CLI don't execute all queries sometimes.

# Install the CLI

```
# Clone this repo
$ git clone https://github.com/perylemke/timebench.git

# Running make command to up database
$ make up

# Running make command to run migrations
$ make migrate

# Running make command to build a binary of CLI
$ make build

# Export env var of URI
$ export DB_CONN_URI='postgres://postgres:password@localhost:5432/homework'
```

# Using

```
# Call CLI and pass a CSV file to a parameter
$ ./bin/timebench start --file /path/to/you/file.csv
Starting queries on DB. Awaiting...
Nice. All queries executed...

Now, show the statistics...

Total queries: 2000
Total time (Seconds): 32.37942078900004
Minimum time (Seconds): 0.010760386
Maximum time (Seconds): 0.032483573
Mean time (Seconds): 0.01618971039450002
Median time (Seconds): 0.015804742
```

# Clean installation
```
# Running make commands to clean
$ make down
$ make clean
```
