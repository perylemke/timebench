package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"
)

func runBenchmarking(cmd *cobra.Command, args []string) {
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

	// Open file
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Unable to open file: %v\n", err)
	}

	// Check if file is valid
	fExt := filepath.Ext(fileName)
	if fExt != ".csv" {
		log.Fatalf("Invalid file extension: %v\n", fExt)
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
	var times []float64
	count := 0
	for {
		// Load line by line on the memory.
		// It's better to not overloading the hardware.
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

			tStart := time.Now()
			_, err = conn.Query(ctx, queryCpuUsage)
			tFinish := time.Since(tStart)

			if err != nil {
				log.Fatalf("Unable to execute query %v\n", err)
			}
			<-guard
			count += 1
			times = append(times, tFinish.Seconds()) // Convert to seconds and append on array
		}(record[0], record[1], record[2])
	}
	wg.Wait()

	// Sort the times.
	sort.Float64s(times)

	// Friendly messages
	fmt.Println("Nice. All queries executed...\n")
	fmt.Println("Now, show the statistics...\n")

	// Show the total of queries
	fmt.Printf("Total queries: %v\n", count)
	time.Sleep(1 * time.Second)

	// Calculate the sum of times.
	sumQueries := 0.0
	for _, s := range times {
		sumQueries += s
	}
	// Print total time
	fmt.Printf("Total time (Seconds): %v\n", sumQueries)
	time.Sleep(1 * time.Second)

	// Print the minimum time in Seconds
	fmt.Printf("Minimum time (Seconds): %v\n", times[0])
	time.Sleep(1 * time.Second)

	// Print the maximum time in Seconds
	fmt.Printf("Maximum time (Seconds): %v\n", times[len(times)-1])
	time.Sleep(1 * time.Second)

	// Calculate mean time and print
	total := 0.0
	for _, v := range times {
		total += v
	}
	fmt.Printf("Mean time (Seconds): %v\n", total/float64(len(times)))
	time.Sleep(1 * time.Second)

	// Calculate Median and print
	if len(times)%2 == 0 {
		fmt.Printf("Median time (Seconds): %v\n", times[len(times)/2])
	} else {
		fmt.Printf("Median time (Seconds): %v\n", (times[(len(times)/2)-1]+times[len(times)/2])/2)
	}

}

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A command to start benchmarking",
	Long:  `A command to start benchmarking in TimescaleDB.`,
	Run:   runBenchmarking,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("file", "", "Input csv file to return the statistics")
}
