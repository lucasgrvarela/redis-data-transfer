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
2023/07/24 10:49:29 Starting the Redis data transfer application...
   8% |█████                                                                | (875/10000, 2 it/s) [8m42s:1h36m15s]
```