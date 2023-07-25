package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/schollz/progressbar/v3"
)

func main() {
	// Define flags to receive Redis connection strings from the command-line arguments
	var oldRedisURL, newRedisURL string
	flag.StringVar(&oldRedisURL, "old", "", "Old Redis connection string (e.g., '[user]:[pass]@url:port')")
	flag.StringVar(&newRedisURL, "new", "", "New Redis connection string (e.g., '[user]:[pass]@url:port')")
	flag.Parse()

	// Ensure both old and new Redis connection strings are provided
	if oldRedisURL == "" || newRedisURL == "" {
		fmt.Println("Please provide the old and new Redis connection strings using the -old and -new flags.")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the logger
	log.Println("Starting the Redis data transfer application...")

	// Create Redis clients for the old and new Redis instances
	oldClient := redis.NewClient(&redis.Options{
		Addr: oldRedisURL,
		DB:   0,
	})

	newClient := redis.NewClient(&redis.Options{
		Addr: newRedisURL,
		DB:   0,
	})

	// Get the number of keys in the old Redis database
	oldKeysCount, err := oldClient.DBSize(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error getting the count of keys from the old Redis database: %v", err)
	}

	// Flush all data from the new Redis database
	_ = newClient.FlushAll(context.Background())

	// Get all keys from the old Redis database
	keys, err := oldClient.Keys(context.Background(), "*").Result()
	if err != nil {
		log.Fatalf("Error getting keys: %v", err)
	}

	// Initialize the progress bar with the total number of keys to transfer
	bar := progressbar.Default(int64(len(keys)))

	// Variables to measure the total time taken for GET and SET operations
	totalGetTime := time.Duration(0)
	totalSetTime := time.Duration(0)

	// Create a Redis pipeline for bulk operations
	oldPipeline := oldClient.Pipeline()
	newPipeline := newClient.Pipeline()

	// Channel to receive keys and communicate between goroutines
	keysChan := make(chan string, len(keys))

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	numWorkers := 100
	wg.Add(numWorkers)

	// Start goroutines for GET and SET operations
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for key := range keysChan {
				if key == "" {
					return // Exit the goroutine when signaled
				}

				// Measure time taken for GET operation
				startTime := time.Now()
				dumpData, err := oldClient.Dump(context.Background(), key).Result()
				if err != nil {
					log.Fatalf("Error getting dump data for key %s: %v", key, err)
				}
				getTime := time.Since(startTime)
				totalGetTime += getTime

				// Add the SET operation to the pipeline
				startTime = time.Now()
				newPipeline.Restore(context.Background(), key, 0, dumpData)
				setTime := time.Since(startTime)
				totalSetTime += setTime

				// Increment the progress bar to show the key transfer progress
				bar.Add(1)
			}
		}()
	}

	// Add keys to the channel to be processed by goroutines
	for _, key := range keys {
		keysChan <- key
	}

	// Close the keysChan channel to signal that all keys have been added
	close(keysChan)

	// Wait for all goroutines to finish before executing the pipelines
	wg.Wait()

	// Execute the pipelines
	_, err = oldPipeline.Exec(context.Background())
	if err != nil {
		log.Fatalf("Error executing the old Redis pipeline: %v", err)
	}

	_, err = newPipeline.Exec(context.Background())
	if err != nil {
		log.Fatalf("Error executing the new Redis pipeline: %v", err)
	}

	// Get the number of keys in the new Redis database
	newKeysCount, err := newClient.DBSize(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error getting the count of keys from the new Redis database: %v", err)
	}

	// Print the transfer summary
	log.Println("Data transfer completed successfully.")
	log.Printf("Number of keys in the old Redis database: %d", oldKeysCount)
	log.Printf("Number of keys in the new Redis database: %d", newKeysCount)
	log.Printf("Total time taken for GET operations: %v", totalGetTime)
	log.Printf("Total time taken for SET operations: %v", totalSetTime)
}
