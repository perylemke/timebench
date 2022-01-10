# Timebench

A small CLI to generate statistics of TimescaleDB.

# Tools

- Go
- Docker
- Docker Compose

# Install

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
Show the statistics...
Query: SELECT time_bucket($1, ts) as time,  max(usage) as max_usage, min(usage) as min_usage 
				FROM cpu_usage
				WHERE host = $2 AND ts BETWEEN $3::timestamp AND $4::timestamp 
				GROUP BY time
Calls: 200
Total time (seconds): 0.4794795470000003
Minimum Time (seconds): 0.001171763
Average Time (seconds): 0.0023973977349999987
Maximum Time (Seconds): 0.004390066
```
