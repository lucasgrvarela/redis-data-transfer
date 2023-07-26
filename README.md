# Redis Data Transfer

This Go program allows you to transfer data between two Redis instances. I created to use to solve a real problem I had to migrate data from some ElasticCache instances on AWS to other instances on MemoryStore on GCP.

### Usage example locally
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

### Usage example inside Kubernetes
```
1. Build the image and send it to your registry (docker build, docker push) 

2. Replace the fields: image, -old and -new inside job.yaml

3. kubectl apply -f job.yaml

4. kubectl logs -f job/redis-data-transfer-job
2023/07/26 12:47:54 Starting the Redis data transfer application...
 100% |██████████████████████████████████| (310406/310406, 1641 it/s)          
2023/07/26 12:51:05 Data transfer completed successfully.
2023/07/26 12:51:05 Number of keys in the old Redis database: 310406
2023/07/26 12:51:05 Number of keys in the new Redis database: 310406
```

##### Observations
-  You need to have network connectivity between the two instances you want to transfer data, you have options like creating a tunnel if they are in the cloud and if you want to run locally. You can run inside a Kubernetes cluster if the cluster has connectivy with both instance and there is a lot of others options you can try.
- Based on my tests the performance is around 2k transfer per second, to achieve it you need a good internet connection if running locally.
- I transfered 310k keys in 3 minutes between one instance on AWS and another on GCP, they had Peering connection established between them.