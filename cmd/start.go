package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"
)

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A command to start benchmarking",
	Long:  `A command to start benchmarking in TimescaleDB.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to DB
		ctx := context.Background()
		connStr := os.Getenv("DB_CONN_URI")
		dbpool, err := pgxpool.Connect(ctx, connStr)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}
		defer dbpool.Close()

		// Receive the csv file
		fileName, _ := cmd.Flags().GetString("file")

		// Check if file is valid
		fExt := filepath.Ext(fileName)
		if fExt != ".csv" {
			log.Fatalf("Invalid file extension: %v\n", fExt)
		}

		// Open file
		f, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("Unable to open file: %v\n", err)
		}

		// Close file
		defer f.Close()

		r := csv.NewReader(f)

		// Skip first line
		if _, err := r.Read(); err != nil {
			log.Fatalf("Unable to read file: %v\n", err)
		}

		var wg sync.WaitGroup
		maxGoroutines := 10
		guard := make(chan struct{}, maxGoroutines)

		// Execute queries on Database
		fmt.Println("Starting queries on DB. Awaiting...")
		for {
			record, err := r.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}

			wg.Add(1)
			guard <- struct{}{}
			go func(hostname, startTime, endTime string) {
				defer wg.Done()

				conn, err := dbpool.Acquire(ctx)
				if err != nil {
					log.Fatalf("Unable to acquire connection: %v\n", err)
				}
				defer conn.Release()

				queryCpuUsage := fmt.Sprintf(`
				SELECT 
					time_bucket('1 min', ts) as time,  
					max(usage) as max_usage, 
					min(usage) as min_usage 
				FROM 
					cpu_usage
				WHERE 
					host = '%s' AND 
					ts BETWEEN '%s'::timestamp AND 
					'%s'::timestamp 
				GROUP BY 
					time;`, hostname, startTime, endTime)

				_, err = conn.Query(ctx, queryCpuUsage)
				if err != nil {
					log.Fatalf("Unable to execute query %v\n", err)
				}
				<-guard
			}(record[0], record[1], record[2])
		}
		wg.Wait()

		fmt.Println("Show the statistics...")

		// Run query to show in terminal
		conn, err := dbpool.Acquire(ctx)
		if err != nil {
			log.Fatalf("Unable to acquire connection: %v\n", err)
		}
		defer conn.Release()

		const queryStmts = `
		SELECT 
    		query, 
    		calls, 
    		( total_exec_time / 1000 ) as total,
    		( min_exec_time / 1000) as min_time,
    		( max_exec_time / 1000) as max_time,
    		( mean_exec_time / 1000 ) as avg_time,
    		(((min_exec_time/1000) + (max_exec_time/1000)) / 2) as median
		FROM
    		pg_stat_statements
		WHERE
    		query ilike '%cpu_usage%';
		`
		rows, err := conn.Query(ctx, queryStmts)
		if err != nil {
			log.Fatalf("Unable to acquire connection: %v\n", err)
		}
		defer rows.Close()

		type queryResult struct {
			Query   string
			Calls   int
			Total   float64
			minTime float64
			avgTime float64
			maxTime float64
			median  float64
		}

		var queryResults []queryResult
		for rows.Next() {
			var qr queryResult
			err = rows.Scan(&qr.Query, &qr.Calls, &qr.Total, &qr.minTime, &qr.avgTime, &qr.maxTime, &qr.median)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to scan %v\n", err)
				os.Exit(1)
			}
			queryResults = append(queryResults, qr)
			fmt.Printf("Query: %v\n", qr.Query)
			time.Sleep(1 * time.Second)
			fmt.Printf("Calls: %v\n", qr.Calls)
			time.Sleep(1 * time.Second)
			fmt.Printf("Total time (Seconds): %v\n", qr.Total)
			time.Sleep(1 * time.Second)
			fmt.Printf("Minimum Time (Seconds): %v\n", qr.minTime)
			time.Sleep(1 * time.Second)
			fmt.Printf("Maximum Time (Seconds): %v\n", qr.maxTime)
			time.Sleep(1 * time.Second)
			fmt.Printf("Average Time (Seconds): %v\n", qr.avgTime)
			time.Sleep(1 * time.Second)
			fmt.Printf("Median Time (Seconds): %v\n", qr.median)
			time.Sleep(1 * time.Second)
		}

		if rows.Err() != nil {
			fmt.Fprintf(os.Stderr, "rows Error: %v\n", rows.Err())
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("file", "", "Input csv file to return the statistics")
}
