# Redis Data Transfer

This Go program allows you to transfer data between two Redis instances.

### Usage example
1. Build the binary
```
go build
```

2. Run it
```
$ ./redis-data-transfer -old redis-url-123.abc:6379 -new localhost:6379
Data transfer completed successfully.
Number of keys in the old Redis database: 5043
Number of keys in the new Redis database: 5046
Total time taken for GET operations: 1m20.805863921s
Total time taken for SET operations: 2m57.330858314s
```