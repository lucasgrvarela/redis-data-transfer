package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
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

	oldClient := redis.NewClient(&redis.Options{
		Addr: oldRedisURL,
		DB:   0,
	})

	newClient := redis.NewClient(&redis.Options{
		Addr: newRedisURL,
		DB:   0,
	})

	oldKeysCount, err := oldClient.DBSize(context.Background()).Result()
	if err != nil {
		fmt.Println("Error getting the count of keys from the old Redis database:", err)
		return
	}

	_ = newClient.FlushAll(context.Background())

	keys, err := oldClient.Keys(context.Background(), "*").Result()
	if err != nil {
		fmt.Println("Error getting keys:", err)
		return
	}

	for _, key := range keys {
		dumpData, err := oldClient.Dump(context.Background(), key).Result()
		if err != nil {
			fmt.Println("Error getting dump data for key:", key, err)
			return
		}

		err = newClient.Restore(context.Background(), key, 0, dumpData).Err()
		if err != nil {
			fmt.Println("Error restoring data for key:", key, err)
			return
		}
	}

	newKeysCount, err := newClient.DBSize(context.Background()).Result()
	if err != nil {
		fmt.Println("Error getting the count of keys from the new Redis database:", err)
		return
	}

	fmt.Println("Data transfer completed successfully.")
	fmt.Println("Number of keys in the old Redis database:", oldKeysCount)
	fmt.Println("Number of keys in the new Redis database:", newKeysCount)
}
