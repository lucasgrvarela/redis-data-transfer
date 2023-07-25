# Redis Data Transfer

This Go program allows you to transfer data between two Redis instances.

### Usage example
1. Build the binary
```
go build
```

2. Run it
```
$ redis-cli keys "*"
(empty array)

$ ./redis-data-transfer -old redis-url-123.abc:6379 -new localhost:6379
2023/07/25 13:04:25 Starting the Redis data transfer application...
 100% |███████████████████████████████████████████████| (10000/10000, 469 it/s)         
2023/07/25 13:04:50 Data transfer completed successfully.
2023/07/25 13:04:50 Number of keys in the old Redis database: 10000
2023/07/25 13:04:50 Number of keys in the new Redis database: 10000

$ redis-cli keys "*"
    1) "key2657"
    2) "key8365"
    3) "key6011"
    4) "key4759"
    5) "key9559"
        ...
```