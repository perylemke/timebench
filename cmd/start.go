package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

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
		connStr := "postgres://postgres:password@localhost:5432/homework"
		dbpool, err := pgxpool.Connect(ctx, connStr)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}
		defer dbpool.Close()

		// Receive the csv file
		fileName, _ := cmd.Flags().GetString("file")

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

				queryCPUUsage := fmt.Sprintf(`
				SELECT time_bucket('1 min', ts) as time,  max(usage) as max_usage, min(usage) as min_usage 
				FROM cpu_usage
				WHERE host = '%s' AND ts BETWEEN '%s'::timestamp AND '%s'::timestamp 
				GROUP BY time;`, hostname, startTime, endTime)

				_, err = conn.Query(ctx, queryCPUUsage)
				if err != nil {
					log.Fatalf("Unable to execute query %v\n", err)
				}
				<-guard
			}(record[0], record[1], record[2])
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("file", "", "Input csv file to return the statistics")
}
