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
	var oldRedisURL, newRedisURL string
	flag.StringVar(&oldRedisURL, "old", "", "Old Redis connection string (e.g., '[user]:[pass]@url:port')")
	flag.StringVar(&newRedisURL, "new", "", "New Redis connection string (e.g., '[user]:[pass]@url:port')")
	flag.Parse()

	if oldRedisURL == "" || newRedisURL == "" {
		fmt.Println("Please provide the old and new Redis connection strings using the -old and -new flags.")
		flag.Usage()
		os.Exit(1)
	}

	log.Println("Starting the Redis data transfer application...")

	oldClient := redis.NewClient(&redis.Options{
		Addr:         oldRedisURL,
		DB:           0,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	})

	newClient := redis.NewClient(&redis.Options{
		Addr:         newRedisURL,
		DB:           0,
		ReadTimeout:  10 * time.Minute, // this will define the WriteTimeout too
		WriteTimeout: 10 * time.Minute,
	})

	oldKeysCount, err := oldClient.DBSize(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error getting the count of keys from the old Redis database: %v", err)
	}

	log.Printf("Number of keys in the old Redis database (before transfer): %d", oldKeysCount)

	_ = newClient.FlushAll(context.Background())

	keys, err := oldClient.Keys(context.Background(), "*").Result()
	if err != nil {
		log.Fatalf("Error getting keys: %v", err)
	}

	bar := progressbar.Default(int64(len(keys)))

	newPipeline := newClient.Pipeline()

	keysChan := make(chan string, len(keys))

	var wg sync.WaitGroup
	numWorkers := 100
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for key := range keysChan {
				if key == "" {
					return
				}

				exists, err := oldClient.Exists(context.Background(), key).Result()
				if err != nil {
					log.Printf("Error checking existence of key %s: %v\n", key, err)
					continue
				}

				if exists == 1 { // 1 exist
					dumpData, err := oldClient.Dump(context.Background(), key).Result()
					if err != nil {
						log.Printf("Error getting dump data for key %s: %v\n", key, err)
						continue
					}

					ttl, err := oldClient.TTL(context.Background(), key).Result()
					if err != nil {
						log.Printf("Error getting TTL for key %s: %v\n", key, err)
					}

					newPipeline.RestoreReplace(context.Background(), key, ttl, dumpData)
				} else { // 0 not exist
					log.Printf("Key doesn't exist with value, probably it's a (nil) value and type = none, %s\n", key)
				}
				bar.Add(1)
			}
		}()
	}

	for _, key := range keys {
		keysChan <- key
	}

	close(keysChan)
	wg.Wait()

	_, err = newPipeline.Exec(context.Background())
	if err != nil {
		log.Fatalf("Error executing the new Redis pipeline: %v", err)
	}

	newKeysCount, err := newClient.DBSize(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error getting the count of keys from the new Redis database: %v", err)
	}

	log.Println("Data transfer completed successfully.")
	log.Printf("Number of keys in the old Redis database: %d", oldKeysCount)
	log.Printf("Number of keys in the new Redis database: %d", newKeysCount)

	if oldKeysCount != newKeysCount {
		log.Printf("Exiting job with error, oldKeyCount != newKeyscount")
		os.Exit(1)
	}
}
